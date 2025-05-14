package deepseek

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
)

type Handler struct {
	client *deepseek.Client
}

func NewHandler(client *deepseek.Client) *Handler {
	return &Handler{client: client}
}

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
