package fcall

import (
	"testing"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"

	"github.com/stretchr/testify/assert"
)

// MockFunctionCall 用于测试的模拟实现
type MockFunctionCall struct {
	name string
}

func (m *MockFunctionCall) Name() string {
	return m.name
}

func (m *MockFunctionCall) Call(ctx *domain.StreamContext, req Request) (Response, error) {
	return Response{}, nil
}

func TestFunctionCallRegistry(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		registry := NewFunctionCallRegistry()
		assert.NotNil(t, registry)
		assert.NotNil(t, registry.calls)
	})

	t.Run("Register", func(t *testing.T) {
		registry := NewFunctionCallRegistry()

		tests := []struct {
			name    string
			fc      FunctionCall
			wantErr bool
		}{
			{
				name:    "正常注册",
				fc:      &MockFunctionCall{name: "test_function"},
				wantErr: false,
			},
			{
				name:    "注册空名称",
				fc:      &MockFunctionCall{name: ""},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				registry.Register(tt.fc)
			})
		}
	})

	t.Run("Lookup", func(t *testing.T) {
		registry := NewFunctionCallRegistry()

		// 注册一些测试函数
		testFC1 := &MockFunctionCall{name: "function1"}
		testFC2 := &MockFunctionCall{name: "function2"}
		testFC3 := &MockFunctionCall{name: "function3"}

		registry.Register(testFC1)
		registry.Register(testFC2)
		registry.Register(testFC3)

		tests := []struct {
			name     string
			funcName string
			wantFC   FunctionCall
			wantErr  error
		}{
			{
				name:     "查找存在的函数1",
				funcName: "function1",
				wantFC:   testFC1,
				wantErr:  nil,
			},
			{
				name:     "查找存在的函数2",
				funcName: "function2",
				wantFC:   testFC2,
				wantErr:  nil,
			},
			{
				name:     "查找存在的函数3",
				funcName: "function3",
				wantFC:   testFC3,
				wantErr:  nil,
			},
			{
				name:     "查找不存在的函数",
				funcName: "nonexistent",
				wantFC:   nil,
				wantErr:  ErrFunctionCallNotFound,
			},
			{
				name:     "查找空名称",
				funcName: "",
				wantFC:   nil,
				wantErr:  ErrFunctionCallNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fc, err := registry.Lookup(tt.funcName)

				if tt.wantErr != nil {
					assert.Error(t, err)
					assert.Equal(t, tt.wantErr, err)
					assert.Nil(t, fc)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.wantFC, fc)
				}
			})
		}
	})
}
