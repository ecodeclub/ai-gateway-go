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

package render

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
)

// Handler 负责渲染 Prompt
// 注意这里我们假设 SystemPrompt 是不需要渲染的，因为很少会出现 SystemPrompt 里面还搞占位符的事情
type Handler struct {
	Next stream.Handler
	// 一般来说，system prompt 是不会出现占位符等问题的
	// userPromptTplCache syncx.Map[string, *template.Template]
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Stream(ctx *domain.StreamContext) (stream.Response, error) {
	prompt, err := h.renderUserPrompt(ctx)
	if err != nil {
		return stream.Response{}, err
	}
	turn := ctx.Chat.CurrentTurn()
	step := turn.AssistantRun.CurrentStep()
	step.RenderedUserPrompt = prompt
	return h.Next.Stream(ctx)
}

func (h *Handler) renderUserPrompt(ctx *domain.StreamContext) (string, error) {
	// 确保 Vars 已初始化
	if ctx.Chat.Vars == nil {
		ctx.Chat.Vars = make(map[string]any)
	}
	// 将用户输入放入 Vars，供模板使用
	turn := ctx.Chat.CurrentTurn()
	if turn.UserRun != nil {
		ctx.Chat.Vars["Input"] = turn.UserRun.Content
	}

	step := turn.AssistantRun.CurrentStep()
	// 暂时不用缓存，测试的时候我经常会直接修改数据库数据
	name := fmt.Sprintf("user-%d", step.Cfg.ID)
	tpl := template.New(name)
	tpl, err := tpl.Parse(step.Cfg.Prompt)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	err = tpl.Execute(&buffer, ctx.Chat.Vars)
	return buffer.String(), err
}

// ExecuteContext 模板中能使用什么内容，就取决于这里
type ExecuteContext struct {
	// 用户的输入
	Input string
}
