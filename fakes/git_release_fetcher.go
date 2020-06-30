package fakes

import (
	"io"
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
	GetReleaseAssetCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Asset github.ReleaseAsset
		}
		Returns struct {
			ReadCloser io.ReadCloser
			Error      error
		}
		Stub func(github.ReleaseAsset) (io.ReadCloser, error)
	}
	GetReleaseTarballCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Url string
		}
		Returns struct {
			ReadCloser io.ReadCloser
			Error      error
		}
		Stub func(string) (io.ReadCloser, error)
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
func (f *GitReleaseFetcher) GetReleaseAsset(param1 github.ReleaseAsset) (io.ReadCloser, error) {
	f.GetReleaseAssetCall.Lock()
	defer f.GetReleaseAssetCall.Unlock()
	f.GetReleaseAssetCall.CallCount++
	f.GetReleaseAssetCall.Receives.Asset = param1
	if f.GetReleaseAssetCall.Stub != nil {
		return f.GetReleaseAssetCall.Stub(param1)
	}
	return f.GetReleaseAssetCall.Returns.ReadCloser, f.GetReleaseAssetCall.Returns.Error
}
func (f *GitReleaseFetcher) GetReleaseTarball(param1 string) (io.ReadCloser, error) {
	f.GetReleaseTarballCall.Lock()
	defer f.GetReleaseTarballCall.Unlock()
	f.GetReleaseTarballCall.CallCount++
	f.GetReleaseTarballCall.Receives.Url = param1
	if f.GetReleaseTarballCall.Stub != nil {
		return f.GetReleaseTarballCall.Stub(param1)
	}
	return f.GetReleaseTarballCall.Returns.ReadCloser, f.GetReleaseTarballCall.Returns.Error
}
