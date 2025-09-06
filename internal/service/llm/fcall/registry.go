package fcall

import (
	"errors"

	"github.com/ecodeclub/ekit/syncx"
)

var ErrFunctionCallNotFound = errors.New("函数调用未找到")

type Registry struct {
	calls *syncx.Map[string, FunctionCall]
}

func NewFunctionCallRegistry() *Registry {
	return &Registry{
		calls: &syncx.Map[string, FunctionCall]{},
	}
}

// Register 注册对应的funcCall
func (f *Registry) Register(fc FunctionCall) error {
	name := fc.Name()
	f.calls.Store(name, fc)
	return nil
}

// Lookup 按名索引对应的funcCall
func (f *Registry) Lookup(name string) (FunctionCall, error) {
	val, ok := f.calls.Load(name)
	if !ok {
		return nil, ErrFunctionCallNotFound
	}
	return val, nil
}
