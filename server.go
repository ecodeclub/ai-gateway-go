package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/cohesion-org/deepseek-go"
	"io"
	"log"
	"time"
)

type StreamResponse struct {
	Chan   chan string
	stream deepseek.ChatCompletionStream
}

func (s *StreamResponse) Work() error {
	for {
		chunk, err := s.stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
			return nil
		}

		s.Chan <- chunk.Choices[0].Delta.Content
	}
}

func sendReq(token string) (*StreamResponse, error) {
	client := deepseek.NewClient(token)
	request := deepseek.StreamChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	}
	ctx := context.Background()
	resp, err := client.CreateChatCompletionStream(ctx, &request)
	if err != nil {
		log.Fatalf("ChatCompletionStream failed: %v", err)
		return nil, err
	}

	return &StreamResponse{stream: resp, Chan: make(chan string)}, nil
}

func ask(token string) {
	stream, err := sendReq(token)
	if err != nil {
		fmt.Println("请求错误")
		return
	}

	go func() {
		err := stream.Work()
		if err != nil {
			fmt.Println("stream 错误")
		}
	}()

	time.Sleep(time.Second)

	for content := range stream.Chan {
		fmt.Print(content)
	}
}
