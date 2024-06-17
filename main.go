package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/azurity/flow-table/loader"
	"github.com/azurity/flow-table/render"
	"github.com/azurity/flow-table/render/engine"
	"github.com/xuri/excelize/v2"
)

func main() {
	path := flag.String("template", "", "the template xlsx")
	dataPath := flag.String("data", "", "data files directory")
	outPath := flag.String("output", "output.xlsx", "output xlsx file path")
	flag.Parse()

	if strings.ToLower(filepath.Ext(*path)) != ".xlsx" {
		log.Panicln("only support *.xlsx template file")
	}
	if stat, err := os.Stat(*path); err != nil || stat.IsDir() {
		if err != nil {
			log.Panicln(err)
		} else {
			log.Panicln("file cannot be a directory")
		}
	}
	if strings.ToLower(filepath.Ext(*outPath)) != ".xlsx" {
		*outPath += ".xlsx"
	}

	celEngine, _ := engine.NewCelEngine()

	engines := engine.NewMultiEngine(map[string]render.RenderEngine{
		"js":  engine.NewJsEngine(),
		"py":  engine.NewPyEngine(),
		"cel": celEngine,
	}, map[string]string{
		"javascript": "js",
		"ecmascript": "js",
		"es":         "js",
		"python":     "py",
	})

	loader := &loader.DirectoryLoader{
		SubLoaders: []loader.SubLoaderDesc{
			{
				Loader: &loader.SqliteLoader{},
				Tester: func(val string) bool {
					ext := filepath.Ext(val)
					return ext == ".db" || ext == ".sqlite"
				},
			},
			{
				Loader: &loader.XlSXLoader{},
				Tester: func(val string) bool {
					return strings.ToLower(filepath.Ext(val)) == ".xlsx"
				},
			},
			{
				Loader: &loader.CSVLoader{},
				Tester: func(val string) bool {
					return strings.ToLower(filepath.Ext(val)) == ".csv"
				},
			},
		},
	}

	if *dataPath != "" {
		data, err := loader.Load(*dataPath)
		if err != nil {
			log.Panicln(err)
		}
		err = engines.InitData(data)
		if err != nil {
			log.Panicln(err)
		}
	}

	file, err := excelize.OpenFile(*path)
	if err != nil {
		log.Panicln(err)
	}
	defer file.Close()
	err = render.Render(file, engines)
	if err != nil {
		log.Panicln(err)
	}
	err = file.SaveAs(*outPath)
	if err != nil {
		log.Panicln(err)
	}
	log.Println("[finish]")
}
