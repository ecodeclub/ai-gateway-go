package fcall

import (
	"errors"

	"github.com/ecodeclub/ekit/syncx"
)

var ErrFunctionCallNotFound = errors.New("function call not found")

type FunctionCallRegistry struct {
	fcalls *syncx.Map[string, FunctionCall]
}

func NewFunctionCallRegistry() *FunctionCallRegistry {
	return &FunctionCallRegistry{
		fcalls: &syncx.Map[string, FunctionCall]{},
	}
}

// Lookup 按名索引对应的fcall
func (f *FunctionCallRegistry) Lookup(name string) (FunctionCall, error) {
	val, ok := f.fcalls.Load(name)
	if !ok {
		return nil, ErrFunctionCallNotFound
	}
	return val, nil
}

// Register 注册对应的fcall
func (f *FunctionCallRegistry) Register(fc FunctionCall) error {
	name := fc.Name()
	f.fcalls.Store(name, fc)
	return nil
}
