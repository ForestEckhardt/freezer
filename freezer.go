package freezer

import (
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"
)

type CacheManager struct {
	CacheDir string
	Cache    CacheDB

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

func (c *CacheManager) Load() error {
	file, err := os.Open(c.cacheDBPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	err = gob.NewDecoder(file).Decode(&c.Cache)
	if err != nil {
		return err
	}

	return nil
}

func (c CacheManager) Save() error {
	file, err := os.OpenFile(c.cacheDBPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	err = gob.NewEncoder(file).Encode(&c.Cache)
	if err != nil {
		return err
	}

	return nil
}
