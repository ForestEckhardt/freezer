package fakes

import "sync"

type JamPackager struct {
	ExecuteCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Args []string
		}
		Returns struct {
			Error error
		}
		Stub func([]string) error
	}
}

func (f *JamPackager) Execute(param1 []string) error {
	f.ExecuteCall.Lock()
	defer f.ExecuteCall.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.Args = param1
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1)
	}
	return f.ExecuteCall.Returns.Error
}
