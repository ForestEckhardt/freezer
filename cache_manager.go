package freezer

import (
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"
)

type CacheManager struct {
	Cache CacheDB

	cacheDir string
	dbFile   *os.File
}

type CacheDB map[string]CacheEntry

type CacheEntry struct {
	Version string
	URI     string
}

func NewCacheManager(cacheDir string) CacheManager {
	return CacheManager{
		cacheDir: cacheDir,
	}
}

func (c *CacheManager) Open() error {
	var err error
	_, err = os.Stat(filepath.Join(c.cacheDir, "buildpacks-cache.db"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(c.cacheDir, os.ModePerm)
			if err != nil {
				return err
			}
			c.dbFile, err = os.Create(filepath.Join(c.cacheDir, "buildpacks-cache.db"))
			if err != nil {
				return err
			}
			c.Cache = CacheDB{}
			return nil
		}
		return err
	}

	loadFile, err := os.Open(filepath.Join(c.cacheDir, "buildpacks-cache.db"))
	if err != nil {
		return err
	}

	err = gob.NewDecoder(loadFile).Decode(&c.Cache)
	if err != nil {
		return err
	}

	c.dbFile, err = os.OpenFile(filepath.Join(c.cacheDir, "buildpacks-cache.db"), os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	return nil
}

func (c CacheManager) Close() error {
	err := gob.NewEncoder(c.dbFile).Encode(&c.Cache)
	if err != nil {
		return err
	}
	defer c.dbFile.Close()

	return nil
}

//This function exists for two reasons  one is so that is could have a standard
//getter setter interface and the setter is a more complex function the other is
//to allow for table locking if this were to be adapted for parallel package management
func (c CacheManager) Get(key string) (CacheEntry, bool, error) {
	entry, ok := c.Cache[key]

	if ok {
		_, err := os.Stat(entry.URI)
		if err != nil {
			if os.IsNotExist(err) {
				return entry, !ok, nil
			}
			return CacheEntry{}, !ok, err
		}
	}

	return entry, ok, nil
}

func (c *CacheManager) Set(key string, value CacheEntry) error {
	//os.RemoveAll of a empty string is a noop if the entry does not exist then it will
	//return and empty string
	err := os.RemoveAll(c.Cache[key].URI)
	if err != nil {
		return err
	}

	if c.Cache == nil {
		return errors.New("the cache manager is not loaded properly")
	}

	c.Cache[key] = value

	return nil
}

func (c CacheManager) Dir() string {
	return c.cacheDir
}
