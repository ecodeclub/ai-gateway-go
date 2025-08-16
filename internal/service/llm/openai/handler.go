package openai

import (
	"context"
	"encoding/json"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/packages/ssestream"
)

type Delta struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}

type Handler struct {
	client openai.Client
	logger *elog.Component
	model  string
}

func NewHandler(apikey string, baseURL string, model string) *Handler {
	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apikey),
	)
	return &Handler{
		client: client,
		logger: elog.DefaultLogger,
		model:  model,
	}
}

func (h *Handler) Chat(ctx context.Context, messages []domain.Message) (domain.ChatResponse, error) {
	params := openai.ChatCompletionNewParams{
		Messages: h.toOpenAIMessage(messages),
		Model:    h.model,
	}
	res, err := h.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return domain.ChatResponse{}, err
	}
	return domain.ChatResponse{
		Response: domain.Message{
			Content: res.Choices[0].Message.Content,
		},
	}, err
}

func (h *Handler) StreamHandle(ctx context.Context, req []domain.Message) (chan domain.StreamEvent, error) {
	eventCh := make(chan domain.StreamEvent, 10)
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(req[0].Content),
		},

		Model: h.model,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: param.Opt[bool]{Value: true},
		},
	}

	go func() {
		stream := h.client.Chat.Completions.NewStreaming(ctx, params)
		h.recv(eventCh, stream)
	}()

	return eventCh, nil
}

func (h *Handler) recv(eventCh chan domain.StreamEvent,
	stream *ssestream.Stream[openai.ChatCompletionChunk]) {
	defer close(eventCh)
	acc := openai.ChatCompletionAccumulator{}

	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		// 建议在处理完 JustFinished 事件后使用数据块
		if len(chunk.Choices) > 0 {
			// 说明没结束
			var delta Delta
			err := json.Unmarshal([]byte(chunk.Choices[0].Delta.RawJSON()), &delta)
			if err != nil {
				eventCh <- domain.StreamEvent{
					Error: err,
				}
				return
			}
			eventCh <- domain.StreamEvent{
				Content:          delta.Content,
				ReasoningContent: delta.ReasoningContent,
			}
		}
	}
	eventCh <- domain.StreamEvent{
		Done: true,
	}
}

func (h *Handler) toOpenAIMessage(messages []domain.Message) []openai.ChatCompletionMessageParamUnion {
	return slice.Map(messages, func(idx int, src domain.Message) openai.ChatCompletionMessageParamUnion {
		switch src.Role {
		case domain.USER:
			return openai.UserMessage(src.Content)
		case domain.SYSTEM:
			return openai.SystemMessage(src.Content)
		default:
			return openai.ChatCompletionMessageParamUnion{}
		}
	})
}
