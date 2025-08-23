//go:build unit

package emit_json

import (
	"encoding/json"
	"testing"

	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"

	"github.com/stretchr/testify/assert"
)

func TestEmitJsonFunctionCall_Call(t *testing.T) {
	tests := []struct {
		name    string
		args    []byte
		ctxData map[string]any
		wantErr bool
	}{
		{
			name:    "正常JSON对象",
			args:    []byte(`{"key": "value"}`),
			ctxData: map[string]any{},
			wantErr: false,
		},
		{
			name:    "正常JSON数组",
			args:    []byte(`[1, 2, 3]`),
			ctxData: map[string]any{},
			wantErr: false,
		},
		{
			name:    "包含前缀的JSON",
			args:    []byte(`prefix {"key": "value"}`),
			ctxData: map[string]any{},
			wantErr: false,
		},
		{
			name:    "包含后缀的JSON",
			args:    []byte(`{"key": "value"} suffix`),
			ctxData: map[string]any{},
			wantErr: false,
		},
		{
			name:    "包含前缀和后缀的JSON",
			args:    []byte(`prefix {"key": "value"} suffix`),
			ctxData: map[string]any{},
			wantErr: false,
		},
		{
			name:    "空输入",
			args:    []byte{},
			ctxData: map[string]any{},
			wantErr: true,
		},
		{
			name:    "无JSON内容",
			args:    []byte(`no json here`),
			ctxData: map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := EmitJsonFunctionCall{}
			ctx := &fcall.Context{Data: tt.ctxData}
			req := fcall.Request{Args: tt.args}

			resp, err := fc.Call(ctx, req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, fcall.Response{}, resp)

			// 验证Context中的数据是否正确设置
			ctxMap := ctx.Data

			extractedJSON, exists := ctxMap["emit_json"]
			assert.True(t, exists)

			// 验证提取的JSON是否有效
			jsonBytes, ok := extractedJSON.([]byte)
			assert.True(t, ok)
			assert.True(t, len(jsonBytes) > 0)

			// 添加反序列化过程，验证返回的json可以反序列化
			var result any
			err = json.Unmarshal(jsonBytes, &result)
			assert.NoError(t, err, "提取的JSON应该能够正常反序列化")

		})
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []byte
		wantErr bool
	}{
		{
			name:    "纯JSON对象",
			input:   []byte(`{"key": "value"}`),
			want:    []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "纯JSON数组",
			input:   []byte(`[1, 2, 3]`),
			want:    []byte(`[1, 2, 3]`),
			wantErr: false,
		},
		{
			name:    "前缀+JSON对象",
			input:   []byte(`prefix {"key": "value"}`),
			want:    []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "前缀+JSON数组",
			input:   []byte(`prefix [1, 2, 3]`),
			want:    []byte(`[1, 2, 3]`),
			wantErr: false,
		},
		{
			name:    "JSON对象+后缀",
			input:   []byte(`{"key": "value"} suffix`),
			want:    []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "JSON数组+后缀",
			input:   []byte(`[1, 2, 3] suffix`),
			want:    []byte(`[1, 2, 3]`),
			wantErr: false,
		},
		{
			name:    "前缀+JSON+后缀",
			input:   []byte(`prefix {"key": "value"} suffix`),
			want:    []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "嵌套JSON",
			input:   []byte(`prefix {"key": {"nested": "value"}} suffix`),
			want:    []byte(`{"key": {"nested": "value"}}`),
			wantErr: false,
		},
		{
			name:    "复杂JSON数组",
			input:   []byte(`prefix [{"id": 1}, {"id": 2}] suffix`),
			want:    []byte(`[{"id": 1}, {"id": 2}]`),
			wantErr: false,
		},
		{
			name:    "空输入",
			input:   []byte{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "无JSON内容",
			input:   []byte(`no json here`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "只有空格",
			input:   []byte(`   `),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "只有特殊字符",
			input:   []byte(`!@#$%^&*()`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "JSON在中间",
			input:   []byte(`start {"key": "value"} end`),
			want:    []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "多个JSON，取第一个",
			input:   []byte(`{"first": "json"} {"second": "json"}`),
			want:    []byte(`{"first": "json"}`),
			wantErr: false,
		},
		{
			name:    "数组优先于对象",
			input:   []byte(`[1, 2, 3] {"key": "value"}`),
			want:    []byte(`[1, 2, 3]`),
			wantErr: false,
		},
		{
			name:    "对象优先于数组（当对象在前时）",
			input:   []byte(`{"key": "value"} [1, 2, 3]`),
			want:    []byte(`{"key": "value"}`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractJSON(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrJsonNotFound, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractJSON_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "不完整的JSON对象",
			input:   []byte(`{"key": "value"`),
			wantErr: true,
		},
		{
			name:    "不完整的JSON数组",
			input:   []byte(`[1, 2, 3`),
			wantErr: true,
		},
		{
			name:    "无效的JSON语法",
			input:   []byte(`{"key": value}`),
			wantErr: true,
		},
		{
			name:    "只有开始符号",
			input:   []byte(`{`),
			wantErr: true,
		},
		{
			name:    "只有开始符号",
			input:   []byte(`[`),
			wantErr: true,
		},
		{
			name:    "只有结束符号",
			input:   []byte(`}`),
			wantErr: true,
		},
		{
			name:    "只有结束符号",
			input:   []byte(`]`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractJSON(tt.input)
			assert.Error(t, err)
			assert.Equal(t, ErrJsonNotFound, err)
		})
	}
}

func TestExtractJSON_Unicode(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []byte
		wantErr bool
	}{
		{
			name:    "包含中文的JSON",
			input:   []byte(`prefix {"message": "你好世界"} suffix`),
			want:    []byte(`{"message": "你好世界"}`),
			wantErr: false,
		},
		{
			name:    "包含emoji的JSON",
			input:   []byte(`prefix {"emoji": "🚀🎉✨"} suffix`),
			want:    []byte(`{"emoji": "🚀🎉✨"}`),
			wantErr: false,
		},
		{
			name:    "包含特殊Unicode字符",
			input:   []byte(`prefix {"unicode": "café résumé naïve"} suffix`),
			want:    []byte(`{"unicode": "café résumé naïve"}`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractJSON(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
