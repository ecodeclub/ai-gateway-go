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

package domain

import (
	"context"
	"time"

	ai "github.com/ecodeclub/ai-gateway-go/api/proto/gen/chat/v1"
)

type Chat struct {
	Sn    string
	Uid   int64
	Title string

	// Vars 跨 step 传递的变量
	// 该字段和 Attachments 的区别是这个是变量
	// 小心 any 类型中确切类型是数字的问题
	Vars map[string]any

	// 业务有关的编排信息
	Orchestration Orchestration

	Biz BizConfig

	Turns []*Turn
	Ctime time.Time
}

type Orchestration struct {
	// Threads 存储的是 Thread 有关的上下文
	Threads map[string]*Thread `json:"threads"`

	// Main 不是主要的意思，而是 main 函数的那个 main，万物起点
	Main *Thread `json:"main"`
}

// LLMConversation 代表第三方的 Conversation
type LLMConversation struct {
	ID string
}

// CurrentTurn 返回当前这一轮
// 它不会检查是否存在
func (c *Chat) CurrentTurn() *Turn {
	return c.Turns[len(c.Turns)-1]
}

func (c *Chat) CurrentStep() *Step {
	return c.CurrentTurn().AssistantRun.CurrentStep()
}

func (c *Chat) History() []HistoryRecord {
	res := make([]HistoryRecord, 0, len(c.Turns)*2)
	for _, t := range c.Turns {
		res = append(res, HistoryRecord{
			Content: t.UserRun.Content,
			Role:    ai.RoleUser,
		})
		// 一般是因为 AI 出了问题
		if t.AssistantRun.Content != "" {
			res = append(res, HistoryRecord{
				Content: t.AssistantRun.Content,
				Role:    ai.RoleAssistant,
			})
		}
	}
	return res
}

type Turn struct {
	ID           int64
	UserRun      *UserRun
	AssistantRun *AssistantRun
}

// AssistantRun 代表的是AI执行的内容
type AssistantRun struct {
	// 这是返回的具体内容
	Content string

	// 很多步骤，每个步骤可能是 LLM 调用，可能是 function call
	// 每次 LLM 调用，都需要一个 Prompt + 大模型返回的数据
	Steps []*Step
}

func (a *AssistantRun) StartLLMStep(thread *Thread) {
	a.Steps = append(a.Steps, &Step{
		Thread: thread,
	})
}

// CurrentStep 返回最后一个步骤
func (a *AssistantRun) CurrentStep() *Step {
	return a.Steps[len(a.Steps)-1]
}

// UserRun 代表的是用户的输入
type UserRun struct {
	// 用户输入的内容
	Content string
	// Files 中存储的是文件地址
	Files []string
}

type Step struct {
	// 渲染好的 User Prompt
	RenderedUserPrompt string
	// 配置
	Cfg InvocationConfigVersion

	Thread *Thread
}

type UserInput = ai.UserInput

type StreamContext struct {
	Ctx    context.Context
	Sender StreamEventSender
	Chat   Chat
}

// History 对话的上下文，要注意的是后续可能会有灵活的策略来决定什么才是需要发送到 LLM 的上下文
func (ctx *StreamContext) History() []HistoryRecord {
	return ctx.Chat.History()
}

type HistoryRecord struct {
	Role    string
	Content string
}

// StreamEventSender 用于和 proto 中的 stream server 解耦
type StreamEventSender interface {
	Send(evt StreamEvent) error
}

type StreamEvent struct {
	// 纯文本
	Delta *Delta
	Usage *Usage
	// 执行步骤更新的信息，主要用于提醒用户
	StepUpdate *StepUpdate
	// Err 出现了错误
	Err error
}

type StepUpdate struct {
	// 步骤的名字
	Title string
	// 步骤当前的执行状态
	Summary string
}

// Usage 记录了大模型的消耗
// 但是要小心，不同的大模型可能会少一两个字段
type Usage struct {
	InputTokens  int64
	OutputTokens int64

	TotalTokens int64
}

type Delta struct {
	Content string
}
