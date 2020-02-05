package fakes

import (
	"context"
	"net/http"
	"sync"

	"github.com/google/go-github/github"
)

type GitClient struct {
	DoCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Ctx context.Context
			Req *http.Request
			V   interface {
			}
		}
		Returns struct {
			Response *github.Response
			Error    error
		}
		Stub func(context.Context, *http.Request, interface {
		}) (*github.Response, error)
	}
}

func (f *GitClient) Do(param1 context.Context, param2 *http.Request, param3 interface {
}) (*github.Response, error) {
	f.DoCall.Lock()
	defer f.DoCall.Unlock()
	f.DoCall.CallCount++
	f.DoCall.Receives.Ctx = param1
	f.DoCall.Receives.Req = param2
	f.DoCall.Receives.V = param3
	if f.DoCall.Stub != nil {
		return f.DoCall.Stub(param1, param2, param3)
	}
	return f.DoCall.Returns.Response, f.DoCall.Returns.Error
}
