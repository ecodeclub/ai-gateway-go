package service

import (
	"context"
	"errors"
	"github.com/cohesion-org/deepseek-go"
	"io"
)

type AIService struct {
	token string
}

func NewAIService(token string) *AIService {
	return &AIService{token: token}
}

func (svc *AIService) Ask(id int64, content string) (*StreamResponse, error) {
	client := deepseek.NewClient(svc.token)
	request := deepseek.StreamChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleUser,
				Content: content,
			},
		},
		Stream: true,
	}
	ctx := context.Background()
	resp, err := client.CreateChatCompletionStream(ctx, &request)

	if err != nil {
		return nil, err
	}

	return &StreamResponse{stream: resp, Chan: make(chan string)}, nil
}

type StreamResponse struct {
	Chan   chan string
	stream deepseek.ChatCompletionStream
}

func (s *StreamResponse) Work() error {
	for {
		chunk, err := s.stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				s.Chan <- "finished"
				return nil
			}
			return err
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
			s.Chan <- "finished"
			return nil
		}

		s.Chan <- chunk.Choices[0].Delta.Content
	}
}
