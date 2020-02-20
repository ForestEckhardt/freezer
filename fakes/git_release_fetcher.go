package fakes

import (
	"sync"

	"github.com/ForestEckhardt/freezer/github"
)

type GitReleaseFetcher struct {
	GetCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Org  string
			Repo string
		}
		Returns struct {
			Release github.Release
			Error   error
		}
		Stub func(string, string) (github.Release, error)
	}
}

func (f *GitReleaseFetcher) Get(param1 string, param2 string) (github.Release, error) {
	f.GetCall.Lock()
	defer f.GetCall.Unlock()
	f.GetCall.CallCount++
	f.GetCall.Receives.Org = param1
	f.GetCall.Receives.Repo = param2
	if f.GetCall.Stub != nil {
		return f.GetCall.Stub(param1, param2)
	}
	return f.GetCall.Returns.Release, f.GetCall.Returns.Error
}
