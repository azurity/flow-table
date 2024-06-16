package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/azurity/flow-table/render"
	"github.com/dop251/goja"
)

type JsEngine struct {
	vm *goja.Runtime
}

func NewJsEngine() *JsEngine {
	return &JsEngine{
		vm: goja.New(),
	}
}

func (engine *JsEngine) InitData(data map[string]any) error {
	for key, value := range data {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		_, err = engine.vm.RunString(fmt.Sprintf("const %s = %s;", key, string(data)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (engine *JsEngine) CalcValue(formula *render.FlowFormula) (data [][]any, rows int, cols int, err error) {
	val, err := engine.vm.RunString(formula.Code)
	if err != nil {
		return nil, 0, 0, err
	}
	switch formula.Direct {
	case render.FlowFormulaDirect_Table:
		data = engine.extractValue(val, 2, formula.Format.Type).([][]any)
		if len(data) == 0 {
			data = append(data, []any{})
		}
	case render.FlowFormulaDirect_H:
		val := engine.extractValue(val, 1, formula.Format.Type).([]any)
		data = [][]any{val}
	case render.FlowFormulaDirect_V:
		val := engine.extractValue(val, 1, formula.Format.Type).([]any)
		data = [][]any{}
		for _, item := range val {
			data = append(data, []any{item})
		}
		if len(data) == 0 {
			data = append(data, []any{})
		}
	case render.FlowFormulaDirect_Cell:
		val := engine.extractValue(val, 0, formula.Format.Type)
		data = [][]any{{val}}
	}

	rows = len(data)
	cols = 1
	for _, row := range data {
		if len(row) > cols {
			cols = len(row)
		}
	}
	for r, row := range data {
		for i := len(row); i < cols; i++ {
			data[r] = append(data[r], paddingValue(formula.Format.Type))
		}
	}
	return data, rows, cols, nil
}

func (engine *JsEngine) extractValue(value goja.Value, level int, class string) any {
	if level == 0 {
		switch class {
		case render.FlowFormulaFormat_String:
			return value.ToString().String()
		case render.FlowFormulaFormat_Int:
			t := value.ExportType()
			if t.Kind() == reflect.Float64 || t.Kind() == reflect.Int64 {
				return int(value.ToInteger())
			} else {
				return 0
			}
		default:
			t := value.ExportType()
			if t.Kind() == reflect.Float64 || t.Kind() == reflect.Int64 {
				return float64(value.ToFloat())
			} else {
				return math.NaN()
			}
		}
	} else {
		ret := []any{}
		kind := value.ExportType().Kind()
		if kind == reflect.Slice || kind == reflect.Array {
			object := value.ToObject(engine.vm)
			length := len(object.Export().([]any))
			for i := 0; i < length; i++ {
				name := strconv.FormatInt(int64(i), 10)
				ret = append(ret, engine.extractValue(object.Get(name), level-1, class))
			}
		} else {
			ret = append(ret, engine.extractValue(value, level-1, class))
		}
		return ret
	}
}
