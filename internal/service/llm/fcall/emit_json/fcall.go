package emit_json

import (
	"encoding/json"

	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"
	"github.com/pkg/errors"
)

const jsonDataName = "data"

var ErrJsonNotFound = errors.New("没找到json")

type EmitJsonFunctionCall struct {
}

func (e EmitJsonFunctionCall) Name() string {
	return "emit_json"
}

// Call 这里的会将
func (e EmitJsonFunctionCall) Call(fctx *fcall.Context, req fcall.Request) (fcall.Response, error) {
	// 第一步从req的data中获取原始数据
	val, err := req.GetVal(jsonDataName)
	if err != nil {
		return fcall.Response{}, err
	}
	jsonData := make(map[string]any)
	err = json.Unmarshal([]byte(val), &jsonData)
	fctx.JSONData = jsonData
	return fcall.Response{}, err
}
