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

package transcriber_test

import (
	"os"
	"testing"

	"github.com/ecodeclub/ai-gateway-go/internal/pkg/transcriber"
	"github.com/stretchr/testify/assert"
)

func TestTranscriber(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Fatal("未设置 API_KEY 环境变量")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		t.Fatal("未设置 BASE_URL 环境变量")
	}

	transcriber := transcriber.NewQwen3ASRFlash(baseURL, apiKey)
	texts, err := transcriber.Transcribe([]string{
		"https://dashscope.oss-cn-beijing.aliyuncs.com/audios/welcome.mp3",
		"https://dashscope.oss-cn-beijing.aliyuncs.com/audios/welcome.mp3",
	})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"欢迎与使用阿里云。", "欢迎与使用阿里云。"}, texts)
}
