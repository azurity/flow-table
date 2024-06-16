package loader

import (
	"errors"
	"os"
	"path/filepath"
)

type LoaderTypeTest func(val string) bool

type SubLoaderDesc struct {
	Loader Loader
	Tester LoaderTypeTest
}

type DirectoryLoader struct {
	SubLoaders []SubLoaderDesc
}

func (loader *DirectoryLoader) Simple() bool {
	return false
}

func (loader *DirectoryLoader) Load(val string) (map[string]any, error) {
	if stat, err := os.Stat(val); err != nil || !stat.IsDir() {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("not a directory")
	}
	entries, err := os.ReadDir(val)
	if err != nil {
		return nil, err
	}

	ret := map[string]any{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		varName := entry.Name()
		varName = varName[:len(varName)-len(filepath.Ext(entry.Name()))]
		subVal := filepath.Join(val, entry.Name())
		for _, loader := range loader.SubLoaders {
			if !loader.Tester(subVal) {
				continue
			}
			loaded, err := loader.Loader.Load(subVal)
			if err != nil {
				return nil, err
			}
			if loader.Loader.Simple() && len(loaded) == 1 {
				for _, value := range loaded {
					ret[varName] = value
				}
			} else {
				ret[varName] = loaded
			}
		}
	}
	return ret, nil
}
