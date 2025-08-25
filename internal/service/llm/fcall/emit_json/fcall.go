package emit_json

import (
	"encoding/json"
	"fmt"

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
	dataMap := make(map[string]any)
	err := json.Unmarshal(req.Args, &dataMap)
	if err != nil {
		return fcall.Response{}, fcall.NewFcallErr(e, fmt.Errorf("反序列化失败 %w", err))
	}
	jsonData, ok := dataMap[jsonDataName]
	if !ok {
		return fcall.Response{}, fcall.NewFcallErr(e, ErrJsonNotFound)
	}
	// 将用户输入的内容放进ctx里
	fctx.JSONData, ok = jsonData.(string)
	if !ok {
		return fcall.Response{}, fcall.NewFcallErr(e, errors.New("jsonData的数据类型不正确"))
	}
	return fcall.Response{}, err
}
