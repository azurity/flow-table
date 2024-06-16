package render

import (
	"strings"

	"github.com/xuri/excelize/v2"
)

type Area struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

func NewArea(dim string) (*Area, error) {
	dimCells := strings.Split(dim, ":")
	if len(dimCells) == 1 {
		dimCells = append(dimCells, dimCells[0])
	}
	left, top, err := excelize.CellNameToCoordinates(dimCells[0])
	if err != nil {
		return nil, err
	}
	right, bottom, err := excelize.CellNameToCoordinates(dimCells[1])
	if err != nil {
		return nil, err
	}
	return &Area{
		Left:   left,
		Top:    top,
		Right:  right,
		Bottom: bottom,
	}, nil
}

func (area Area) Contains(col int, row int) bool {
	return col >= area.Left && col <= area.Right && row >= area.Top && row <= area.Bottom
}

func Render(workbook *excelize.File, engine RenderEngine) error {
	for _, sheet := range workbook.GetSheetList() {

		dim, err := workbook.GetSheetDimension(sheet)
		if err != nil {
			return err
		}
		area, err := NewArea(dim)
		if err != nil {
			return err
		}

		calcMergeArea := func() []*Area {
			mergeCells, err := workbook.GetMergeCells(sheet)
			if err != nil {
				return nil
			}
			mergeAreas := []*Area{}
			for _, mCell := range mergeCells {
				area, err := NewArea(mCell.GetStartAxis() + ":" + mCell.GetEndAxis())
				if err != nil {
					return nil
				}
				mergeAreas = append(mergeAreas, area)
			}
			return mergeAreas
		}

		mergeAreas := calcMergeArea()

		for currentRow := area.Top; currentRow <= area.Bottom; currentRow += 1 {
			for currentCol := area.Left; currentCol <= area.Right; currentCol += 1 {
				inMerge := false
				for _, mArea := range mergeAreas {
					if mArea.Contains(currentCol, currentRow) {
						inMerge = true
						break
					}
				}
				if inMerge {
					continue
				}

				cellName, _ := excelize.CoordinatesToCellName(currentCol, currentRow)
				cellType, _ := workbook.GetCellType(sheet, cellName)

				if cellType != excelize.CellTypeInlineString && cellType != excelize.CellTypeSharedString {
					continue
				}
				value, _ := workbook.GetCellValue(sheet, cellName)
				formula := TryParseFlowFormula(value)
				if formula == nil {
					continue
				}

				rendered, rows, cols, err := engine.CalcValue(formula)
				if err != nil {
					return err
				}

				if rows > 1 {
					workbook.InsertRows(sheet, currentRow+1, rows-1)
				}
				if cols > 1 {
					col, _ := excelize.ColumnNumberToName(currentCol + 1)
					workbook.InsertCols(sheet, col, cols-1)
				}
				for r := 0; r < rows; r++ {
					for c := 0; c < cols; c++ {
						newCellName, _ := excelize.CoordinatesToCellName(currentCol+c, currentRow+r)
						switch formula.Format.Type {
						case FlowFormulaFormat_String:
							if val, ok := rendered[r][c].(string); ok {
								err := workbook.SetCellStr(sheet, newCellName, val)
								if err != nil {
									return err
								}
							}
						case FlowFormulaFormat_Int:
							if val, ok := rendered[r][c].(int); ok {
								err := workbook.SetCellInt(sheet, newCellName, val)
								if err != nil {
									return err
								}
							}
						case FlowFormulaFormat_Float:
							if val, ok := rendered[r][c].(float64); ok {
								err := workbook.SetCellFloat(sheet, newCellName, val, formula.Format.Constraint, 64)
								if err != nil {
									return err
								}
							}
						case FlowFormulaFormat_Percent:
							if val, ok := rendered[r][c].(float64); ok {
								constraint := formula.Format.Constraint
								if constraint > 0 {
									constraint += 2
								}
								err := workbook.SetCellFloat(sheet, newCellName, val, constraint, 64)
								if err != nil {
									return err
								}
							}
						}
					}
				}

				styleId, err := workbook.GetCellStyle(sheet, cellName)
				if err != nil {
					return err
				}
				style, err := workbook.GetStyle(styleId)
				if err != nil {
					return err
				}
				newStyle, err := workbook.NewStyle(&excelize.Style{
					Border:        style.Border,
					Fill:          style.Fill,
					Font:          style.Font,
					Alignment:     style.Alignment,
					Protection:    style.Protection,
					NumFmt:        0,
					DecimalPlaces: style.DecimalPlaces,
					CustomNumFmt:  formula.Format.GenerateFormatStr(),
				})
				if err != nil {
					return err
				}
				areaCell, _ := excelize.CoordinatesToCellName(currentCol+cols-1, currentRow+rows-1)
				workbook.SetCellStyle(sheet, cellName, areaCell, newStyle)

				area.Right += cols - 1
				area.Bottom += rows - 1
				mergeAreas = calcMergeArea()
			}
		}
	}
	return nil
}
