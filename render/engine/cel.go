package engine

import (
	"log"
	"reflect"
	"regexp"

	"github.com/azurity/flow-table/render"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
)

type CelEngine struct {
	env  *cel.Env
	data map[string]any
}

var jsonRegExp = regexp.MustCompile(`(^|\b)json:"(.+)"($|\b)`)

func fieldJsonName(t reflect.StructField) string {
	if !jsonRegExp.MatchString(string(t.Tag)) {
		return t.Name
	}
	match := jsonRegExp.FindStringSubmatch(string(t.Tag))
	return match[2]
}

func struct2Map(obj reflect.Value) any {
	t := obj.Type()

	switch t.Kind() {
	case reflect.Struct:
		data := map[string]any{}
		for i := 0; i < t.NumField(); i++ {
			data[fieldJsonName(t.Field(i))] = struct2Map(obj.Field(i))
		}
		return data
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		data := []any{}
		for i := 0; i < obj.Len(); i++ {
			elem := obj.Index(i)
			if elem.Type().Kind() == reflect.Interface {
				data = append(data, struct2Map(reflect.ValueOf(obj.Index(i).Interface())))
			} else {
				data = append(data, struct2Map(obj.Index(i)))
			}
		}
		return data
	case reflect.Map:
		data := map[string]any{}
		for _, key := range obj.MapKeys() {
			data[key.String()] = struct2Map(reflect.ValueOf(obj.MapIndex(key).Interface()))
		}
		return data
	case reflect.Pointer:
		return struct2Map(obj.Elem())
	default:
		return obj.Addr().Interface()
	}
}

func NewCelEngine() (*CelEngine, error) {
	env, err := cel.NewEnv(
		cel.Variable("data", cel.MapType(cel.StringType, cel.AnyType)),
		// cel.Variable("data", cel.AnyType),
	)
	if err != nil {
		return nil, err
	}
	return &CelEngine{
		env: env,
	}, nil
}

func (engine *CelEngine) InitData(data map[string]any) error {
	engine.data = struct2Map(reflect.ValueOf(data)).(map[string]any)
	return nil
}

func (engine *CelEngine) CalcValue(formula *render.FlowFormula) (data [][]any, rows int, cols int, err error) {
	ast, issue := engine.env.Compile(formula.Code)
	if issue.Err() != nil {
		return nil, 0, 0, issue.Err()
	}
	program, err := engine.env.Program(ast)
	if err != nil {
		return nil, 0, 0, issue.Err()
	}
	value, _, err := program.Eval(map[string]any{
		"data": engine.data,
	})
	if err != nil {
		log.Println(err)
		return nil, 0, 0, err
	}

	switch formula.Direct {
	case render.FlowFormulaDirect_Table:
		val, err := engine.extractValue(value, 2, formula.Format.Type)
		if err != nil {
			return nil, 0, 0, err
		}
		data = val.([][]any)
		if len(data) == 0 {
			data = append(data, []any{})
		}
	case render.FlowFormulaDirect_H:
		val, err := engine.extractValue(value, 1, formula.Format.Type)
		if err != nil {
			return nil, 0, 0, err
		}
		data = [][]any{val.([]any)}
	case render.FlowFormulaDirect_V:
		val, err := engine.extractValue(value, 1, formula.Format.Type)
		if err != nil {
			return nil, 0, 0, err
		}
		data = [][]any{}
		for _, item := range val.([]any) {
			data = append(data, []any{item})
		}
		if len(data) == 0 {
			data = append(data, []any{})
		}
	case render.FlowFormulaDirect_Cell:
		val, err := engine.extractValue(value, 0, formula.Format.Type)
		if err != nil {
			return nil, 0, 0, err
		}
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

func (engine *CelEngine) extractValue(value ref.Val, level int, class string) (any, error) {
	var elemType reflect.Type
	switch class {
	case render.FlowFormulaFormat_String:
		elemType = reflect.TypeOf("")
	case render.FlowFormulaFormat_Int:
		elemType = reflect.TypeOf(int(0))
	default:
		elemType = reflect.TypeOf(float64(0))
	}
	if level == 0 {
		return value.ConvertToNative(elemType)
	} else {
		retType := elemType
		for i := 0; i < level; i++ {
			retType = reflect.TypeOf(retType)
		}
		ret, err := value.ConvertToNative(retType)
		if err == nil {
			return ret, nil
		}
		ret, err = engine.extractValue(value, level-1, class)
		if err == nil {
			return []any{ret}, nil
		}
		return nil, err
	}
}
