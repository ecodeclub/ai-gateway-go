package service

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm"
)

// AIService 是一个结构体，包含一个 LLMHandler 的指针
type AIService struct {
	handler llm.LLMHandler
}

// NewAIService 创建一个新的 AIService 实例
// 参数 handler 是 llm.LLMHandler 类型，表示处理 LLM 请求的处理器
// 返回值是 *AIService 类型
func NewAIService(handler llm.LLMHandler) *AIService {
	return &AIService{handler: handler}
}

// Stream 处理流式请求
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 req 是 domain.LLMRequest 类型，表示客户端的请求
// 返回值是一个 channel，用于返回 stream 事件，以及可能发生的错误
func (svc *AIService) Stream(ctx context.Context, req domain.LLMRequest) (chan domain.StreamEvent, error) {
	return svc.handler.StreamHandle(ctx, req)
}

// Invoke 处理普通请求
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 req 是 domain.LLMRequest 类型，表示客户端的请求
// 返回值是 domain.LLMResponse 类型，表示处理结果，以及可能发生的错误
func (svc *AIService) Invoke(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error) {
	return svc.handler.Handle(ctx, req)
}
