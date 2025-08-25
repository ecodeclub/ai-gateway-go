//go:build unit

package emit_json

import (
	"errors"
	"testing"

	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitJsonFunctionCall_Name(t *testing.T) {
	fc := EmitJsonFunctionCall{}
	assert.Equal(t, "emit_json", fc.Name())
}

func TestEmitJsonFunctionCall_Call(t *testing.T) {
	tests := []struct {
		name        string
		args        []byte
		wantJSON    string
		wantErr     bool
		errIs       error
		errContains []string
	}{
		{
			name:     "成功",
			args:     []byte(`{"data":"hello"}`),
			wantJSON: "hello",
		},
		{
			name:        "反序列化失败",
			args:        []byte(`not json`),
			wantErr:     true,
			errContains: []string{"反序列化失败"},
		},
		{
			name:    "缺少data",
			args:    []byte(`{"other":"x"}`),
			wantErr: true,
			errIs:   ErrJsonNotFound,
		},
		{
			name:        "data类型不正确",
			args:        []byte(`{"data":123}`),
			wantErr:     true,
			errContains: []string{"jsonData的数据类型不正确"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := EmitJsonFunctionCall{}
			ctx := &fcall.Context{}
			_, err := fc.Call(ctx, fcall.Request{Args: tt.args})

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "functionCall: emit_json")
				for _, s := range tt.errContains {
					assert.ErrorContains(t, err, s)
				}
				if tt.errIs != nil {
					assert.True(t, errors.Is(err, tt.errIs))
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantJSON, ctx.JSONData)
		})
	}
}
