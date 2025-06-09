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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ecodeclub/ai-gateway-go/errs"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
)

// Config 配置选项
type Config struct {
	APIKey     string
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
}

// DefaultConfig 返回默认配置
func DefaultConfig(apiKey string) Config {
	return Config{
		APIKey:     apiKey,
		BaseURL:    "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}
}

type Handler struct {
	config Config
	client *http.Client
}

func NewHandler(config Config) *Handler {
	return &Handler{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// StreamHandle 处理流式请求 api:https://help.aliyun.com/zh/model-studio/qwen-omni
func (h *Handler) StreamHandle(ctx context.Context, llmRequest domain.LLMRequest) (chan domain.StreamEvent, error) {
	events := make(chan domain.StreamEvent, 10)

	go func() {
		defer close(events)
		// 创建可取消的上下文（支持超时控制）
		ctx1, cancel := context.WithTimeout(ctx, h.config.Timeout)
		defer cancel()

		cs := h.buildContent(llmRequest)
		messages := []Messages{{Role: "user", Content: cs}}
		reqBody := NewSendRequestBuilder(messages).Build(WithModalities([]string{"text", "audio"}))

		marshal, err := json.Marshal(reqBody)
		if err != nil {
			events <- domain.StreamEvent{Error: fmt.Errorf("序列化请求失败: %v", err)}
			return
		}

		req, err := h.buildRequest(ctx1, marshal)
		if err != nil {
			events <- domain.StreamEvent{Error: fmt.Errorf("创建请求失败: %v", err)}
			return
		}

		resp, err := h.doRequest(req)
		if err != nil {
			events <- domain.StreamEvent{Error: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()

		h.recv(ctx, events, resp.Body)
	}()

	return events, nil
}

func (h *Handler) buildContent(llmRequest domain.LLMRequest) []Content {
	switch llmRequest.ContentType {
	case domain.ContentTypeImage:
		return []Content{NewImageContent(llmRequest.Text)}
	case domain.ContentTypeAudio:
		return []Content{NewInputAudioContent(llmRequest.Text, "wav")}
	case domain.ContentTypeText:
		return []Content{NewTextContent(llmRequest.Text)}
	case domain.ContentTypeVideo:
		return []Content{NewVideoContent([]string{llmRequest.Text}), NewTextContent("请参考视频内容")}
	default:
		return []Content{NewTextContent(llmRequest.Text)}
	}
}

func (h *Handler) buildRequest(ctx context.Context, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", h.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.config.APIKey))

	return req, nil
}

func (h *Handler) doRequest(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i < h.config.MaxRetries; i++ {
		resp, err = h.client.Do(req)
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				return resp, nil
			}
			_ = resp.Body.Close()
		}
		time.Sleep(time.Second * time.Duration(i+1))
	}

	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	return nil, fmt.Errorf("非200响应: %s", resp.Status)
}

func (h *Handler) recv(ctx context.Context, eventCh chan domain.StreamEvent, stream io.ReadCloser) {
	reader := bufio.NewReader(stream)
	for {
		if ctx.Err() != nil {
			eventCh <- domain.StreamEvent{Error: fmt.Errorf("上下文取消: %v", ctx.Err())}
			return
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				eventCh <- domain.StreamEvent{Done: true}
				return
			}
			eventCh <- domain.StreamEvent{Error: fmt.Errorf("读取错误: %v", err)}
			return
		}

		lineStr := string(line)
		if len(lineStr) <= 6 || !strings.HasPrefix(lineStr, "data: ") {
			eventCh <- domain.StreamEvent{Error: fmt.Errorf("解析数据 %s", lineStr)}
			return
		}

		if strings.Contains(lineStr, "data: [DONE]") {
			eventCh <- domain.StreamEvent{Done: true}
			return
		}

		jsonData := strings.TrimSpace(lineStr[6:])
		var resp StreamResponse
		if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
			eventCh <- domain.StreamEvent{Error: fmt.Errorf("解析JSON错误: %v ,原始数据为 %s", err, lineStr)}
			continue
		}

		if len(resp.Choices) > 0 {
			if resp.Choices[0].Delta.Content != "" {
				eventCh <- domain.StreamEvent{Content: resp.Choices[0].Delta.Content}
			} else if resp.Choices[0].Delta.Audio.Transcript != "" {
				eventCh <- domain.StreamEvent{Content: resp.Choices[0].Delta.Audio.Transcript}
			}
		}
	}
}

func (h *Handler) Handle(_ context.Context, _ domain.LLMRequest) (domain.LLMResponse, error) {
	return domain.LLMResponse{}, errs.ErrApiNotSupport
}
