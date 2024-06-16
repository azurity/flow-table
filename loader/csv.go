package loader

import (
	"encoding/csv"
	"os"
)

type CSVLoader struct{}

func (loader *CSVLoader) Simple() bool {
	return true
}

func (loader *CSVLoader) Load(val string) (map[string]any, error) {
	file, err := os.Open(val)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"data": data,
	}, nil
}
