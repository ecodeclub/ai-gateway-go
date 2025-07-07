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

//go:build onlyUsage

package openai

import (
	"context"
	"os"
	"testing"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/stretchr/testify/require"
)

func TestOpenai(t *testing.T) {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("ALI_QIANWEN_DASHSCOPE_API_KEY")),
		option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1/"),
	)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Say this is a test"),
		},
		Model: "qwen-plus",
	})
	if err != nil {
		panic(err.Error())
	}

	println(chatCompletion.Choices[0].Message.Content)
}

func TestHandle(t *testing.T) {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("ALI_QIANWEN_DASHSCOPE_API_KEY")),
		option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1/"),
	)
	h := NewHandler(&client, "qwen-plus")
	handle, err := h.Handle(context.Background(), []domain.Message{
		{
			Role:    domain.SYSTEM,
			Content: "你好",
		},
		{
			Role:    domain.SYSTEM,
			Content: "你是谁",
		},
	})

	require.NoError(t, err)
	println(handle.Response.Content)
	require.NotEmpty(t, handle.Response.Content)
}

func TestStreamHandle(t *testing.T) {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("ALI_QIANWEN_DASHSCOPE_API_KEY")),
		option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1/"),
	)
	h := NewHandler(&client, "qwen-plus")
	s, err := h.StreamHandle(context.Background(), []domain.Message{
		{
			Role:    domain.SYSTEM,
			Content: "你好",
		},
		{
			Role:    domain.SYSTEM,
			Content: "你是谁",
		},
	})

	require.NoError(t, err)

	for {
		select {
		case event, ok := <-s:
			if !ok {
				return
			}
			if event.Error != nil {
				t.Errorf("stream error: %v", event.Error)
				return
			}
			if event.Done {
				println("stream done")
				return
			}
			println(event.Content)
			require.NotEmpty(t, event.Content)
		}
	}

}
