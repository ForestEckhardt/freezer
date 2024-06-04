package fakes

import "sync"

type Packager struct {
	ExecuteCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			BuildpackDir string
			Output       string
			Version      string
			Cached       bool
		}
		Returns struct {
			Error error
		}
		Stub func(string, string, string, bool) error
	}
}

func (f *Packager) Execute(param1 string, param2 string, param3 string, param4 bool) error {
	f.ExecuteCall.mutex.Lock()
	defer f.ExecuteCall.mutex.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.BuildpackDir = param1
	f.ExecuteCall.Receives.Output = param2
	f.ExecuteCall.Receives.Version = param3
	f.ExecuteCall.Receives.Cached = param4
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1, param2, param3, param4)
	}
	return f.ExecuteCall.Returns.Error
}
