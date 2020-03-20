package fakes

import "sync"

type Packager struct {
	ExecuteCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			BuildpackDir string
			Output       string
			Version      string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string, string) error
	}
}

func (f *Packager) Execute(param1 string, param2 string, param3 string) error {
	f.ExecuteCall.Lock()
	defer f.ExecuteCall.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.BuildpackDir = param1
	f.ExecuteCall.Receives.Output = param2
	f.ExecuteCall.Receives.Version = param3
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1, param2, param3)
	}
	return f.ExecuteCall.Returns.Error
}
