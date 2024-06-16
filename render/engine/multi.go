package engine

import (
	"errors"
	"regexp"
	"strings"

	"github.com/azurity/flow-table/render"
)

type MultiEngine struct {
	Engines map[string]render.RenderEngine
	Alias   map[string]string
}

func NewMultiEngine(engines map[string]render.RenderEngine, alias map[string]string) *MultiEngine {
	ret := &MultiEngine{
		Engines: engines,
		Alias:   alias,
	}
	for name := range engines {
		ret.Alias[name] = name
	}
	return ret
}

func (engine *MultiEngine) InitData(data map[string]any) error {
	for _, impl := range engine.Engines {
		if err := impl.InitData(data); err != nil {
			return err
		}
	}
	return nil
}

var langRegExp = regexp.MustCompile(`^\[(\w+)\](.*)$`)

var ErrWrongCodeFormat = errors.New("wrong code format")
var ErrUnknownLang = errors.New("unknown language")

func (engine *MultiEngine) CalcValue(formula *render.FlowFormula) (data [][]any, rows int, cols int, err error) {
	if !langRegExp.MatchString(formula.Code) {
		return nil, 0, 0, ErrWrongCodeFormat
	}
	match := langRegExp.FindStringSubmatch(formula.Code)
	lang, ok := engine.Alias[strings.ToLower(match[1])]
	if !ok {
		return nil, 0, 0, ErrUnknownLang
	}
	impl, ok := engine.Engines[lang]
	if !ok {
		return nil, 0, 0, ErrUnknownLang
	}
	formula.Code = match[2]
	return impl.CalcValue(formula)
}
