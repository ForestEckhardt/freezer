package fakes

import (
	"sync"

	"github.com/ForestEckhardt/freezer"
)

type BuildpackCache struct {
	DirCall struct {
		mutex     sync.Mutex
		CallCount int
		Returns   struct {
			String string
		}
		Stub func() string
	}
	GetCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Key string
		}
		Returns struct {
			CacheEntry freezer.CacheEntry
			Bool       bool
			Error      error
		}
		Stub func(string) (freezer.CacheEntry, bool, error)
	}
	SetCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Key         string
			CachedEntry freezer.CacheEntry
		}
		Returns struct {
			Error error
		}
		Stub func(string, freezer.CacheEntry) error
	}
}

func (f *BuildpackCache) Dir() string {
	f.DirCall.mutex.Lock()
	defer f.DirCall.mutex.Unlock()
	f.DirCall.CallCount++
	if f.DirCall.Stub != nil {
		return f.DirCall.Stub()
	}
	return f.DirCall.Returns.String
}
func (f *BuildpackCache) Get(param1 string) (freezer.CacheEntry, bool, error) {
	f.GetCall.mutex.Lock()
	defer f.GetCall.mutex.Unlock()
	f.GetCall.CallCount++
	f.GetCall.Receives.Key = param1
	if f.GetCall.Stub != nil {
		return f.GetCall.Stub(param1)
	}
	return f.GetCall.Returns.CacheEntry, f.GetCall.Returns.Bool, f.GetCall.Returns.Error
}
func (f *BuildpackCache) Set(param1 string, param2 freezer.CacheEntry) error {
	f.SetCall.mutex.Lock()
	defer f.SetCall.mutex.Unlock()
	f.SetCall.CallCount++
	f.SetCall.Receives.Key = param1
	f.SetCall.Receives.CachedEntry = param2
	if f.SetCall.Stub != nil {
		return f.SetCall.Stub(param1, param2)
	}
	return f.SetCall.Returns.Error
}
