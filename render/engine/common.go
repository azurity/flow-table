package engine

import (
	"math"

	"github.com/azurity/flow-table/render"
)

func paddingValue(class string) any {
	switch class {
	case render.FlowFormulaFormat_String:
		return ""
	case render.FlowFormulaFormat_Int:
		return int(0)
	default:
		return math.NaN()
	}
}
