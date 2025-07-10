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

package deepseek

import (
	"context"
	"errors"
	"github.com/cohesion-org/deepseek-go"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ekit/slice"
	"io"
)

type Handler struct {
	client *deepseek.Client
}

func NewHandler(client *deepseek.Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) Handle(ctx context.Context, req []domain.Message) (domain.ChatResponse, error) {
	request := &deepseek.ChatCompletionRequest{
		Model:    deepseek.DeepSeekChat,
		Messages: h.ToMessage(req),
	}
	response, err := h.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return domain.ChatResponse{}, err
	}

	message := domain.Message{
		Role:             response.Choices[0].Message.Role,
		Content:          response.Choices[0].Message.Content,
		ReasoningContent: response.Choices[0].Message.ReasoningContent,
	}

	return domain.ChatResponse{Response: message}, nil
}

func (h *Handler) StreamHandle(ctx context.Context, req []domain.Message) (chan domain.StreamEvent, error) {
	request := deepseek.StreamChatCompletionRequest{
		Model:    deepseek.DeepSeekChat,
		Messages: h.ToMessage(req),
		Stream:   true,
	}

	stream, err := h.client.CreateChatCompletionStream(ctx, &request)
	if err != nil {
		return nil, err
	}
	events := make(chan domain.StreamEvent, 10)
	go func() {
		defer close(events)
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
		eventCh <- domain.StreamEvent{Content: chunk.Choices[0].Delta.Content, ReasoningContent: chunk.Choices[0].Delta.ReasoningContent, Error: nil}
	}
}

func (h *Handler) ToMessage(messages []domain.Message) []deepseek.ChatCompletionMessage {
	return slice.Map(messages, func(idx int, src domain.Message) deepseek.ChatCompletionMessage {
		return deepseek.ChatCompletionMessage{
			Role:    src.Role,
			Content: src.Content,
		}
	})
}
