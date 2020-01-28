package freezer

import (
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"
)

type CacheManager struct {
	CacheDir    string
	cacheDBPath string
}

type CacheDB map[string]CacheEntry

type CacheEntry struct {
	Version string
	URI     string
}

func NewCacheManager(cacheDir string) CacheManager {
	return CacheManager{
		CacheDir:    cacheDir,
		cacheDBPath: filepath.Join(cacheDir, "buildpacks-cache.db"),
	}
}

func (c CacheManager) Load() (CacheDB, error) {
	file, err := os.Open(c.cacheDBPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CacheDB{}, nil
		}
		panic(err)
	}
	defer file.Close()

	var cacheDB CacheDB
	gobDecoder := gob.NewDecoder(file)
	err = gobDecoder.Decode(&cacheDB)
	if err != nil {
		panic(err)
	}

	return cacheDB, nil
}
