package loader

type Loader interface {
	Simple() bool
	Load(val string) (map[string]any, error)
}
