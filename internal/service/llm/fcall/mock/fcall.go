//go:build mock

package mock

import "github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"

type FunctionCall struct {
}

func (m FunctionCall) Name() string {
	return "mock"
}

func (m FunctionCall) Call(ctx *fcall.Context, req fcall.Request) (fcall.Response, error) {
	// 如果说你需要将处理的结果给其他functioncall使用可以放在ctx 的data中
	data := ctx.Data
	data["mock"] = "mockData"
	// 如果说你需要通知客户端（前端）前面是需要前端需要做的functioncall，对应的值是参数
	ctx.Attachments["mockFunc"] = `{"key":"val"}`
	return fcall.Response{}, nil
}
