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
	"github.com/ecodeclub/ekit/mapx"
)

type ChatV1 struct {
	Sn    string
	Uid   int64
	Title string

	// 第三方的 conversation id
	// 目前只有 openai 需要使用
	LLMConversation LLMConversation

	// Vars 跨 step 传递的变量
	// 该字段和 Attachments 的区别是这个是变量
	// 小心 any 类型中确切类型是数字的问题
	Vars  map[string]any
	Turns []*Turn
	Ctime time.Time
}

// LLMConversation 代表第三方的 Conversation
type LLMConversation struct {
	ID string
}

// LastTurn 返回最后一轮，也就是“当前”对话
// 它不会检查是否存在
func (c *ChatV1) LastTurn() *Turn {
	return c.Turns[len(c.Turns)-1]
}

// CombinedVars 返回合并后的变量，也就是将 Turn 和全局的合并
func (c *ChatV1) CombinedVars() map[string]any {
	turn := c.LastTurn()
	return mapx.Merge(c.Vars, turn.Vars)
}

func (c *ChatV1) History() []HistoryRecord {
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
	ID int64
	// 当前这一个轮次产生的变量，会覆盖掉全局的 Vars
	Vars         map[string]any
	UserRun      *UserRun
	AssistantRun *AssistantRun
}

// AssistantRun 代表的是AI执行的内容
type AssistantRun struct {
	// 这是返回的具体内容
	Content string

	Steps []*Step

	// 目前认为 Attachments 就是 key-value 结构
	// 如果 Assistant 本身产出了一些文件等，
	// 那么这里要么是文件的内容，
	// 要么是文件存储到 COS 之后的地址
	Attachments map[string]string
}

func (a *AssistantRun) StartLLMStep(cfgID int64) {
	a.Steps = append(a.Steps, &Step{
		Type: StepTypeFunctionCall,
		Data: &StepLLMData{
			CfgID: cfgID,
		},
	})
}

// LastStep 返回最后一个步骤
func (a *AssistantRun) LastStep() *Step {
	return a.Steps[len(a.Steps)-1]
}

// UserRun 代表的是用户的输入
type UserRun struct {
	// 用户输入的内容
	// 当AudioURL不为空时，将 AudioURL 转为文本并拼接在Content后面
	Content  string
	AudioURL string

	// Files 中存储的是文件地址
	Files []string
}

const (
	StepTypeLLM          = "LLM"
	StepTypeFunctionCall = "FunctionCall"
)

type Step struct {
	// 类型
	Type string
	// Data 是该步骤产生的数据、上下文等
	// 部分步骤会操作 ctx 的其它部分
	Data any
}

func (s *Step) LLMData() *StepLLMData {
	return s.Data.(*StepLLMData)
}

func (s *Step) FunctionCallData() *StepFunctionCallData {
	return s.Data.(*StepFunctionCallData)
}

type StepLLMData struct {
	// 渲染好的 User Prompt
	RenderedUserPrompt string
	CfgID              int64
	// 配置
	Cfg InvocationConfigVersion
}

// StepFunctionCallData 这个是指执行 function call 产生的数据
// 如果是大模型返回的执行 function call 的参数，那么在 StepLLMData 中
type StepFunctionCallData struct {
	// function call 的名字
	Name string
	// function call 产生的各种数据，具体是什么类型取决于具体的 function call
	// function call 需要将
	Data any
}

type UserInput = ai.UserInput

type StreamContext struct {
	Ctx    context.Context
	Sender StreamEventSender
	Chat   ChatV1
}

type RouteInfo struct {
	NextInvCfgID int64
}

func (ctx *StreamContext) LLMCid() string {
	return ctx.Chat.LLMConversation.ID
}

// History 对话的上下文，要注意的是后续可能会有灵活的策略来决定什么才是需要发送到 LLM 的上下文
func (ctx *StreamContext) History() []HistoryRecord {
	return ctx.Chat.History()
}

type HistoryRecord struct {
	Role    string
	Content string
	// 一些附属的内容，也可能需要传递给 LLM
	Attachments map[string]string
}

// StreamEventSender 用于和 proto 中的 stream server 解耦
type StreamEventSender interface {
	Send(evt StreamEventV1) error
}

type StreamEventV1 struct {
	// 纯文本
	Delta *Delta
	Usage *Usage
	// 执行步骤更新的信息，主要用于提醒用户
	StepUpdate *StepUpdate
	// MCP 或者 function call 产生的东西
	// 按名索引，不同的 mcp 和 function call 自己决定 key
	// 用户应该知道如何按名索引，以及对应的数据类型
	Attachments map[string]any
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
