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

package qwen_omni_turbo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
)

func TestStreamHandle(t *testing.T) {
	apiKey := os.Getenv("QWEN_API_KEY")
	if apiKey == "" {
		t.Skip("QWEN_API_KEY environment variable not set")
	}

	config := DefaultConfig(apiKey)
	handler := NewHandler(config)

	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()

	ch, err := handler.StreamHandle(timeout, domain.LLMRequest{
		Text:        "你是谁",
		ContentType: ContentTypeText,
	})
	if err != nil {
		t.Fatalf("StreamHandle failed: %v", err)
	}

	for {
		select {
		case <-timeout.Done():
			t.Log("Test timeout reached")
			return
		case event := <-ch:
			if event.Error != nil {
				t.Fatalf("Received error: %v", event.Error)
			}
			if event.Done {
				t.Log("Stream completed successfully")
				return
			}
			t.Logf("Received content: %s", event.Content)
		}
	}
}

func TestStreamHandleWithInvalidAPIKey(t *testing.T) {
	config := DefaultConfig("invalid-api-key")
	handler := NewHandler(config)

	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()

	ch, err := handler.StreamHandle(timeout, domain.LLMRequest{
		Text:        "你是谁",
		ContentType: ContentTypeText,
	})
	if err != nil {
		t.Fatalf("StreamHandle failed: %v", err)
	}

	for {
		select {
		case <-timeout.Done():
			t.Log("Test timeout reached")
			return
		case event := <-ch:
			if event.Error != nil {
				t.Logf("Expected error received: %v", event.Error)
				return
			}
			if event.Done {
				t.Fatal("Stream completed unexpectedly with invalid API key")
			}
		}
	}
}

func TestStreamHandleWithDifferentContentTypes(t *testing.T) {
	apiKey := os.Getenv("QWEN_API_KEY")
	if apiKey == "" {
		t.Skip("QWEN_API_KEY environment variable not set")
	}

	config := DefaultConfig(apiKey)
	handler := NewHandler(config)

	testCases := []struct {
		name        string
		contentType domain.ContentType
		text        string
	}{
		{
			name:        "Image Content",
			contentType: domain.ContentTypeImage,
			text:        "https://help-static-aliyun-doc.aliyuncs.com/file-manage-files/zh-CN/20241022/emyrja/dog_and_girl.jpeg",
		},
		{
			name:        "Audio Content",
			contentType: domain.ContentTypeAudio,
			text:        "https://help-static-aliyun-doc.aliyuncs.com/file-manage-files/zh-CN/20250211/tixcef/cherry.wav",
		},
		{
			name:        "Video Content",
			contentType: domain.ContentTypeVideo,
			text:        "https://www.bilibili.com/video/BV1RH4y1a7Sg?t=8.3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
			defer cancelFunc()

			ch, err := handler.StreamHandle(timeout, domain.LLMRequest{
				Text:        tc.text,
				ContentType: tc.contentType,
			})
			if err != nil {
				t.Fatalf("StreamHandle failed: %v", err)
			}

			for {
				select {
				case <-timeout.Done():
					t.Logf("Test timeout reached for %s", tc.name)
					return
				case event := <-ch:
					if event.Error != nil {
						t.Logf("Received error for %s: %v", tc.name, event.Error)
						return
					}
					if event.Done {
						t.Logf("Stream completed successfully for %s", tc.name)
						return
					}
					t.Logf("Received content for %s: %s", tc.name, event.Content)
				}
			}
		})
	}
}
