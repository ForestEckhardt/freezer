package freezer

import "fmt"

type LocalBuildpack struct {
	Path        string
	Name        string
	UncachedKey string
	CachedKey   string
	Offline     bool
	Version     string
}

func NewLocalBuildpack(path, name string) LocalBuildpack {
	return LocalBuildpack{
		Path:        path,
		Name:        name,
		UncachedKey: fmt.Sprintf("%s", name),
		CachedKey:   fmt.Sprintf("%s:cached", name),
	}
}
