package service

import (
	"context"
	"errors"
	"github.com/cohesion-org/deepseek-go"
	"io"
)

type StreamResponse struct {
	Chan   chan string
	stream deepseek.ChatCompletionStream
}

func (s *StreamResponse) Work() error {
	for {
		chunk, err := s.stream.Recv()
		if err != nil {
			// 结束
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

func SendReq(token string, content string) (*StreamResponse, error) {
	client := deepseek.NewClient(token)
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
