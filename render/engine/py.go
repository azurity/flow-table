package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/azurity/flow-table/render"
	"github.com/go-python/gpython/py"

	_ "github.com/go-python/gpython/stdlib"
)

type PyEngine struct {
	ctx    py.Context
	module *py.Module
}

func NewPyEngine() *PyEngine {
	ctx := py.NewContext(py.DefaultContextOpts())
	module, _ := ctx.ModuleInit(&py.ModuleImpl{
		Info: py.ModuleInfo{
			FileDesc: "",
		},
	})
	return &PyEngine{
		ctx:    ctx,
		module: module,
	}
}

func (engine *PyEngine) InitData(data map[string]any) error {
	for key, value := range data {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		code, err := py.Compile(fmt.Sprintf("%s = %s\n", key, string(data)), "", py.SingleMode, 0, true)
		if err != nil {
			return err
		}
		_, err = engine.ctx.RunCode(code, engine.module.Globals, engine.module.Globals, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (engine *PyEngine) CalcValue(formula *render.FlowFormula) (data [][]any, rows int, cols int, err error) {
	code, err := py.Compile(formula.Code, "", py.EvalMode, 0, true)
	if err != nil {
		return nil, 0, 0, err
	}
	val, err := engine.ctx.RunCode(code, engine.module.Globals, engine.module.Globals, nil)
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

func (engine *PyEngine) extractValue(value py.Object, level int, class string) any {
	if level == 0 {
		switch class {
		case render.FlowFormulaFormat_String:
			out, err := py.Str(value)
			if err != nil {
				return ""
			}
			str, _ := py.StrAsString(out)
			return str
		case render.FlowFormulaFormat_Int:
			if value.Type() == py.IntType {
				ret, err := value.(py.Int).GoInt()
				if err != nil {
					return 0
				}
				return ret
			} else if value.Type() == py.FloatType {
				ret, err := py.FloatAsFloat64(value)
				if err != nil {
					return 0
				}
				return int(ret)
			}
			return 0
		default:
			if value.Type() == py.FloatType {
				ret, err := py.FloatAsFloat64(value)
				if err != nil {
					return 0
				}
				return ret
			}
			return math.NaN()
		}
	} else {
		ret := []any{}
		if rawLen, err := py.Len(value); err == nil && value.Type() != py.StringType {
			length, err := rawLen.(py.Int).GoInt()
			if err != nil {
				return []any{}
			}
			for i := 0; i < length; i++ {
				index, err := py.IntFromString(strconv.FormatInt(int64(i), 10), 10)
				if err != nil {
					return []any{}
				}
				item, err := py.GetItem(value, index)
				if err != nil {
					return []any{}
				}
				ret = append(ret, engine.extractValue(item, level-1, class))
			}
		} else {
			ret = append(ret, engine.extractValue(value, level-1, class))
		}
		return ret
	}
}
