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

package buildvar

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/pkg/transcriber"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/gotomicro/ego/core/elog"
)

// Handler 构建var
type Handler struct {
	transcriber transcriber.Transcriber
	Next        stream.Handler
	logger      *elog.Component
}

func NewHandler(transcriber transcriber.Transcriber) *Handler {
	return &Handler{
		transcriber: transcriber,
		logger:      elog.DefaultLogger.With(elog.FieldComponentName("buildvar.Handler"))}
}

func (h *Handler) Stream(ctx *domain.StreamContext) error {
	turn := ctx.Chat.LastTurn()
	userInput := turn.UserRun

	// 语音输入转化为文本
	if userInput.AudioURL != "" && h.isAudioFile(userInput.AudioURL) {
		texts, err := h.transcriber.Transcribe([]string{userInput.AudioURL})
		if err != nil {
			h.logger.Error("语音识别为文本失败", elog.FieldErr(err))
			return fmt.Errorf("语音识别为文本失败：%w", err)
		}
		// 将音频识别结果设置为用户输入
		userInput.Content += "\n" + texts[0]
	}

	turn.Vars["Input"] = userInput.Content
	h.logger.Info("build var", elog.String("Input", userInput.Content))
	return h.Next.Stream(ctx)
}

func (h *Handler) isAudioFile(fileURL string) bool {
	audioExtensions := map[string]bool{
		".aac":  true,
		".amr":  true,
		".aiff": true,
		".flac": true,
		".m4a":  true,
		".mp3":  true,
		".mpeg": true,
		".ogg":  true,
		".opus": true,
		".wav":  true,
		".webm": true,
		".wma":  true,
	}
	return audioExtensions[strings.ToLower(filepath.Ext(fileURL))]
}
