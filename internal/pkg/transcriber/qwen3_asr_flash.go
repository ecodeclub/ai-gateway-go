// Copyright 2025 ecodeclub
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

package transcriber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gotomicro/ego/core/elog"
)

// Qwen3ASRFlash 将音频文件转化为文字，当前只处理第一个找到的音频文件
type Qwen3ASRFlash struct {
	baseURL string
	apiKey  string
	logger  *elog.Component
}

func NewQwen3ASRFlash(baseURL, apikey string) *Qwen3ASRFlash {
	return &Qwen3ASRFlash{
		baseURL: baseURL,
		apiKey:  apikey,
		logger:  elog.DefaultLogger.With(elog.FieldComponentName("transcriber.qwen3.Qwen3ASRFlash"))}
}

func (h *Qwen3ASRFlash) Transcribe(fileURLs []string) (texts []string, err error) {
	for i := range fileURLs {
		text, err := h.transcribe(fileURLs[i])
		if err != nil {
			return nil, err
		}
		texts = append(texts, text)
	}
	return texts, nil
}

func (h *Qwen3ASRFlash) transcribe(file string) (text string, err error) {
	req := map[string]any{
		"model": "qwen3-asr-flash", // 支持中文、英文，无需申请即可使用
		"input": map[string]any{
			"messages": []map[string]any{
				{
					"role": "system",
					"content": []map[string]any{
						{
							"text": "请将以下音频文件URL转换为文字：",
						},
					},
				},
				{
					"role": "user",
					"content": []map[string]any{
						{
							"audio": file,
						},
					},
				},
			},
		},
		"parameters": map[string]any{
			"asr_options": map[string]any{
				"enable_itn": true,
			},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败：%w", err)
	}

	url := h.baseURL + "/services/aigc/multimodal-generation/generation"

	h.logger.Info("语音识别请求",
		elog.String("URL", url),
		elog.String("Body", string(reqBody)),
		elog.String("APIKey(前10字符)", h.apiKey[:min(10, len(h.apiKey))]))

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("构建识别请求失败：%w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+h.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 60 * time.Second}
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("发送识别请求失败：%w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("语音识别失败: HTTP %d", httpResp.StatusCode)
	}

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败：%w", err)
	}

	// 解析文本
	type Content struct {
		Text string `json:"text"`
	}
	var output struct {
		Output struct {
			Choices []struct {
				Message struct {
					Content []Content `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		} `json:"output"`
	}

	err = json.Unmarshal(respBody, &output)
	if err != nil {
		return "", fmt.Errorf("解析输出失败: %w, 原始响应: %s", err, string(respBody))
	}
	if len(output.Output.Choices) == 0 || len(output.Output.Choices[0].Message.Content) == 0 {
		return "", fmt.Errorf("解析输出失败: 原始响应: %s", string(respBody))
	}
	return output.Output.Choices[0].Message.Content[0].Text, nil
}
