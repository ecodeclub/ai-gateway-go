package fcall

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmitJsonFunctionCall(t *testing.T) {
	tests := []struct {
		name        string
		argMap      map[string]string
		wantErr     error
		wantJSON    map[string]any
		checkPrefix bool
	}{
		{
			name:     "成功解析JSON到上下文",
			argMap:   map[string]string{"data": `{"a":1,"b":"x"}`},
			wantErr:  nil,
			wantJSON: map[string]any{"a": float64(1), "b": "x"},
		},
		{
			name:    "缺少data参数",
			argMap:  map[string]string{},
			wantErr: ErrArgNotFound,
			// 失败情况下，JSONData 未被写入，应为 nil
			wantJSON: nil,
		},
		{
			name:        "data不是合法JSON",
			argMap:      map[string]string{"data": "not a json"},
			wantErr:     errors.New("json: "),
			wantJSON:    (map[string]any)(nil),
			checkPrefix: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := EmitJsonFunctionCall{}
			fctx := &Context{
				Context: t.Context()}

			args, err := json.Marshal(tt.argMap)
			assert.NoError(t, err)

			resp, err := fc.Call(fctx, Request{Args: args})
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, Response{}, resp)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.wantJSON, fctx.JSONData)
		})
	}
}
