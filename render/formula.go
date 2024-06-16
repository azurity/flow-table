package render

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	FlowFormulaDirect_Cell int = iota
	FlowFormulaDirect_H
	FlowFormulaDirect_V
	FlowFormulaDirect_Table
)

var FlowFormulaDirectName = map[string]int{
	"C": FlowFormulaDirect_Cell,
	"H": FlowFormulaDirect_H,
	"V": FlowFormulaDirect_V,
	"T": FlowFormulaDirect_Table,
}

const (
	FlowFormulaFormat_String  string = "s" // string
	FlowFormulaFormat_Float   string = "f" // float64
	FlowFormulaFormat_Int     string = "d" // int
	FlowFormulaFormat_Percent string = "p" // float64
)

type FlowFormulaFormat struct {
	Type       string
	Constraint int
}

func ParseFlowFormulaFormat(value string) FlowFormulaFormat {
	class := string([]byte{value[len(value)-1]})
	switch class {
	case FlowFormulaFormat_String:
		return FlowFormulaFormat{
			Type: FlowFormulaFormat_String,
		}
	case FlowFormulaFormat_Int:
		var i int64 = -1
		if len(value) > 1 {
			i, _ = strconv.ParseInt(value[:len(value)-1], 10, 64)
		}
		return FlowFormulaFormat{
			Type:       FlowFormulaFormat_Int,
			Constraint: int(i),
		}
	case FlowFormulaFormat_Percent:
		fallthrough
	case FlowFormulaFormat_Float:
		var i int64 = -1
		if len(value) > 2 {
			i, _ = strconv.ParseInt(value[1:len(value)-1], 10, 64)
		}
		return FlowFormulaFormat{
			Type:       class,
			Constraint: int(i),
		}
	}
	return FlowFormulaFormat{}
}

func (format FlowFormulaFormat) GenerateFormatStr() *string {
	ret := "General"
	constraint := format.Constraint
	if constraint < 0 {
		constraint = 2
	}
	switch format.Type {
	case FlowFormulaFormat_Percent:
		ret = "0." + strings.Repeat("0", format.Constraint) + "%"
	case FlowFormulaFormat_Float:
		ret = "0." + strings.Repeat("0", format.Constraint)
	}
	return &ret
}

type FlowFormula struct {
	Direct int
	Format FlowFormulaFormat
	Code   string
}

var formulaRegExp = regexp.MustCompile(`^\{\{((?P<direct>C|H|V|T)(\((?P<format>.+)\))?\|)?(?P<exp>.+)\}\}$`)
var formatRegExp = regexp.MustCompile(`s|((\.\d+)?f|p)|\d*d`)

func TryParseFlowFormula(value string) *FlowFormula {
	if !formulaRegExp.MatchString(value) {
		return nil
	}
	match := formulaRegExp.FindStringSubmatch(value)

	direct, ok := FlowFormulaDirectName[match[formulaRegExp.SubexpIndex("direct")]]
	if !ok {
		direct = FlowFormulaDirect_Cell
	}

	rawFormat := match[formulaRegExp.SubexpIndex("format")]
	if !formatRegExp.MatchString(rawFormat) {
		return nil
	}

	ret := &FlowFormula{
		Direct: direct,
		Format: ParseFlowFormulaFormat(rawFormat),
		Code:   match[formulaRegExp.SubexpIndex("exp")],
	}
	return ret
}

type RenderEngine interface {
	InitData(data map[string]any) error
	CalcValue(formula *FlowFormula) (data [][]any, rows int, cols int, err error)
}
