package loader

import (
	"strings"

	"github.com/xuri/excelize/v2"
)

type XlSXLoader struct{}

func (loader *XlSXLoader) Simple() bool {
	return false
}

func (loader *XlSXLoader) Load(val string) (map[string]any, error) {
	file, err := excelize.OpenFile(val)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sheets := file.GetSheetList()
	ret := map[string]any{}
	for _, sheet := range sheets {
		dim, err := file.GetSheetDimension(sheet)
		if err != nil {
			return nil, err
		}
		dims := strings.Split(dim, ":")
		if len(dims) != 2 {
			continue
		}
		left, top, err := excelize.CellNameToCoordinates(dims[0])
		if err != nil {
			return nil, err
		}
		right, bottom, err := excelize.CellNameToCoordinates(dims[1])
		if err != nil {
			return nil, err
		}
		data := [][]string{}
		for r := top; r <= bottom; r++ {
			row := []string{}
			for c := left; c <= right; c++ {
				cellName, _ := excelize.CoordinatesToCellName(c, r, false)
				value, err := file.GetCellValue(sheet, cellName)
				if err != nil {
					row = append(row, "")
				} else {
					row = append(row, value)
				}
			}
			data = append(data, row)
		}
		ret[sheet] = data
	}
	return ret, nil
}
