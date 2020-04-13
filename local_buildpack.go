package freezer

import "fmt"

type LocalBuildpack struct {
	path        string
	name        string
	uncachedKey string
	cachedKey   string
}

func NewLocalBuildpack(path, name string) LocalBuildpack {
	return LocalBuildpack{
		path:        path,
		name:        name,
		uncachedKey: fmt.Sprintf("%s", name),
		cachedKey:   fmt.Sprintf("%s:cached", name),
	}
}
