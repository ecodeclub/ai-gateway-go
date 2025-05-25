package deepseek

import (
	"context"
	"errors"
	"github.com/cohesion-org/deepseek-go"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"io"
	"time"
)

// Handler 是一个结构体，包含一个 deepseek.Client 的指针
type Handler struct {
	client *deepseek.Client
}

// NewHandler 创建一个新的 Handler 实例
// 参数 client 是 *deepseek.Client 类型，用于与 DeepSeek API 通信
// 返回值是 *Handler 类型
func NewHandler(client *deepseek.Client) *Handler {
	return &Handler{client: client}
}

// Handle 处理普通的 LLM 请求
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 req 是 domain.LLMRequest 类型，表示客户端的请求
// 返回值是 domain.LLMResponse 类型，表示处理结果，以及可能发生的错误
func (h *Handler) Handle(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error) {
	request := &deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleUser,
				Content: req.Text,
			},
		},
	}
	response, err := h.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return domain.LLMResponse{}, err
	}
	return domain.LLMResponse{Content: response.Choices[0].Message.Content}, nil
}

// StreamHandle 处理流式 LLM 请求
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 req 是 domain.LLMRequest 类型，表示客户端的请求
// 返回值是一个 channel，用于返回 stream 事件，以及可能发生的错误
func (h *Handler) StreamHandle(ctx context.Context, req domain.LLMRequest) (chan domain.StreamEvent, error) {
	request := deepseek.StreamChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleUser,
				Content: req.Text,
			},
		},
		Stream: true,
	}
	// 设置对应的 chan
	events := make(chan domain.StreamEvent, 10)

	go func() {
		defer close(events)

		// 设置对应的超时时间
		newCtx, cancel := context.WithTimeout(ctx, time.Minute*10)
		defer cancel()
		stream, err := h.client.CreateChatCompletionStream(newCtx, &request)

		if err != nil {
			events <- domain.StreamEvent{Error: err}
			return
		}

		h.recv(events, stream)
	}()

	return events, nil
}

// recv 方法负责接收并处理来自 DeepSeek 流式响应的数据
// 参数 eventCh 是用于发送事件的 channel
// 参数 stream 是 deepseek.ChatCompletionStream 类型，表示 DeepSeek 的流式响应
func (h *Handler) recv(eventCh chan domain.StreamEvent, stream deepseek.ChatCompletionStream) {
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				eventCh <- domain.StreamEvent{Done: true}
				break
			}
			eventCh <- domain.StreamEvent{Error: err}
		}
		eventCh <- domain.StreamEvent{Content: chunk.Choices[0].Delta.Content, Error: nil}
	}
}
