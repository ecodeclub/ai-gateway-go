package emit_json

import (
	"bytes"
	"encoding/json"

	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"
	"github.com/pkg/errors"
)

var ErrJsonNotFound = errors.New("没找到json")

type EmitJsonFunctionCall struct {
}

func (e EmitJsonFunctionCall) Name() string {
	return "emit_json"
}

// Call 这里的会将
func (e EmitJsonFunctionCall) Call(ctx *fcall.Context, req fcall.Request) (fcall.Response, error) {
	// 第一步从req的data中获取原始数据
	originData := req.Args
	jsonBytes, err := extractJSON(originData)
	if err != nil {
		return fcall.Response{}, err
	}
	ctx.Data[e.Name()] = jsonBytes
	return fcall.Response{}, nil
}

// extractJSON 第三种方式：滑窗 + json.Valid
// 从输入中找到首个 '{' 或 '['，然后逐步扩展窗口，直到 json.Valid 判定为合法 JSON。
func extractJSON(data []byte) ([]byte, error) {
	start := bytes.IndexAny(data, "{")
	alt := bytes.IndexAny(data, "[")
	if start == -1 || (alt != -1 && alt < start) {
		start = alt
	}
	if start < 0 {
		return nil, ErrJsonNotFound
	}
	for end := start + 1; end <= len(data); end++ {
		if json.Valid(data[start:end]) {
			return data[start:end], nil
		}
	}
	return nil, ErrJsonNotFound
}
