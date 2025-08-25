//go:build mock

package mock

import "github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"

type FunctionCall struct {
}

func (m FunctionCall) Name() string {
	return "mock"
}

func (m FunctionCall) Call(ctx *fcall.Context, req fcall.Request) (fcall.Response, error) {
	// Attachments 里存储着所有functionCall的产物
	ctx.Attachments["mockFunc"] = `{"key":"val"}`
	return fcall.Response{}, nil
}
