package llm

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
)

// LLMHandler 是一个接口，定义了处理 LLM 请求的基本方法
type LLMHandler interface {
	// StreamHandle 处理流式 LLM 请求
	// 参数 ctx 是上下文，用于控制请求的生命周期
	// 参数 req 是 domain.LLMRequest 类型，表示客户端的请求
	// 返回值是一个 channel，用于返回 stream 事件，以及可能发生的错误
	StreamHandle(ctx context.Context, req domain.LLMRequest) (chan domain.StreamEvent, error)

	// Handle 处理普通的 LLM 请求
	// 参数 ctx 是上下文，用于控制请求的生命周期
	// 参数 req 是 domain.LLMRequest 类型，表示客户端的请求
	// 返回值是 domain.LLMResponse 类型，表示处理结果，以及可能发生的错误
	Handle(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error)
}
