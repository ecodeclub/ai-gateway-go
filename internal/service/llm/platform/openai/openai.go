// Copyright 2021 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openai

import (
	"context"
	"io"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ekit/slice"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/openai/openai-go/shared/constant"
	"github.com/pkg/errors"
)

type Handler struct {
	client     *openai.Client
	model      string
	toolCallID string
	opts       []option.RequestOption
}

func NewHandler(client *openai.Client, model string, options ...Option) *Handler {
	h := &Handler{
		client: client,
		model:  model,
	}

	for _, o := range options {
		o(h)
	}

	return h
}

type Option func(*Handler)

func WithToolCallID(toolCallID string) Option {
	return func(h *Handler) {
		h.toolCallID = toolCallID
	}
}

func WithRequestOptions(opts ...option.RequestOption) Option {
	return func(h *Handler) {
		h.opts = opts
	}
}

func (h *Handler) Handle(ctx context.Context, req []domain.Message) (domain.ChatResponse, error) {
	messages := h.toMessage(req)
	request := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    h.model,
	}
	response, err := h.client.Chat.Completions.New(ctx, request, h.opts...)
	if err != nil {
		return domain.ChatResponse{}, err
	}

	if len(response.Choices) == 0 {
		return domain.ChatResponse{}, errors.New("no response choices returned")
	}

	message := domain.Message{
		Role:             toDomainRole(response.Choices[0].Message.Role),
		Content:          response.Choices[0].Message.Content,
		ReasoningContent: "", // OpenAI does not provide reasoning content in the response
	}

	return domain.ChatResponse{Response: message}, nil
}

func (h *Handler) StreamHandle(ctx context.Context, req []domain.Message) (chan domain.StreamEvent, error) {
	request := openai.ChatCompletionNewParams{
		Model:    h.model,
		Messages: h.toMessage(req),
	}
	events := make(chan domain.StreamEvent, 10)

	go func() {
		defer close(events)

		// 设置对应的超时时间
		newCtx, cancel := context.WithTimeout(ctx, time.Minute*10)
		defer cancel()
		stream := h.client.Chat.Completions.NewStreaming(newCtx, request, h.opts...)
		if stream.Err() != nil {
			events <- domain.StreamEvent{Error: stream.Err()}
			return
		}

		h.recv(events, stream)
	}()

	return events, nil
}

func (h *Handler) recv(eventCh chan domain.StreamEvent, stream *ssestream.Stream[openai.ChatCompletionChunk]) {
	for stream.Next() {
		if err := stream.Err(); err != nil {
			handleStreamError(eventCh, err)
			return
		}

		node := stream.Current()
		if len(node.Choices) == 0 {
			continue
		}

		choice := node.Choices[0]
		if choice.Delta.Content != "" {
			eventCh <- domain.StreamEvent{Content: choice.Delta.Content, Error: nil}
		}
	}

	// 检查最终错误
	if err := stream.Err(); err != nil {
		handleStreamError(eventCh, err)
	} else {
		eventCh <- domain.StreamEvent{Done: true}
	}
}

func handleStreamError(eventCh chan domain.StreamEvent, err error) {
	if errors.Is(err, io.EOF) {
		eventCh <- domain.StreamEvent{Done: true}
	} else {
		eventCh <- domain.StreamEvent{Error: err}
	}
}

func (h *Handler) toMessage(messages []domain.Message) []openai.ChatCompletionMessageParamUnion {
	return slice.FilterMap(messages, func(idx int, src domain.Message) (openai.ChatCompletionMessageParamUnion, bool) {
		var temp openai.ChatCompletionMessageParamUnion
		switch src.Role {
		case domain.TOOL:
			return openai.ToolMessage(src.Content, h.toolCallID), true
		case domain.SYSTEM:
			return openai.SystemMessage(src.Content), true
		case domain.USER:
			return openai.UserMessage(src.Content), true
		case domain.ASSISTANT:
			return openai.AssistantMessage(src.Content), true
		case domain.UNKNOWN:
			return temp, false
		}
		return temp, false
	})
}

func toDomainRole(c constant.Assistant) int32 {
	switch c {
	case "assistant":
		return domain.ASSISTANT
	default:
		return domain.UNKNOWN
	}
}
