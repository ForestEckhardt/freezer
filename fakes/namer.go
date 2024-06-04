package fakes

import "sync"

type Namer struct {
	RandomNameCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Name string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string) (string, error)
	}
}

func (f *Namer) RandomName(param1 string) (string, error) {
	f.RandomNameCall.mutex.Lock()
	defer f.RandomNameCall.mutex.Unlock()
	f.RandomNameCall.CallCount++
	f.RandomNameCall.Receives.Name = param1
	if f.RandomNameCall.Stub != nil {
		return f.RandomNameCall.Stub(param1)
	}
	return f.RandomNameCall.Returns.String, f.RandomNameCall.Returns.Error
}
