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

package bailian

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/gotomicro/ego/core/elog"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/shared"
)

type Handler struct {
	client   openai.Client
	logger   *elog.Component
	registry *fcall.Registry
	// 发送给百炼需要携带的一些公共的头部
	options []option.RequestOption

	// 发起下一个 LLM 调用
	Handler stream.Handler

	// 维护 Chat.Sn -> Assistant ID 的映射（每个 Chat 对应一个 Assistant）
	assistants *sync.Map
}

func NewHandler(
	client openai.Client,
	registry *fcall.Registry,
	headers map[string]string,
) *Handler {
	headerOpts := make([]option.RequestOption, 0, len(headers))
	for k, v := range headers {
		headerOpts = append(headerOpts, option.WithHeader(k, v))
	}
	return &Handler{
		client:     client,
		registry:   registry,
		options:    headerOpts,
		assistants: &sync.Map{},
		logger:     elog.DefaultLogger.With(elog.FieldComponentName("bailian.Handler")),
	}
}

func (h *Handler) Stream(ctx *domain.StreamContext) error {
	h.logger.Info("🎬 Stream 方法开始执行", elog.String("chatSn", ctx.Chat.Sn))

	// 1. 初始化 Thread
	err := h.initThreadIfNeeded(ctx)
	if err != nil {
		h.logger.Error("❌ 初始化 Thread 失败", elog.FieldErr(err))
		return err
	}
	h.logger.Info("✅ Thread 已初始化", elog.String("threadID", ctx.Chat.LLMConversation.ID))

	// 2. 获取配置和 Assistant
	step := ctx.Chat.LastTurn().AssistantRun.LastStep()
	cfg := step.LLMData().Cfg
	assistantID, err := h.getOrCreateAssistant(ctx, cfg)
	if err != nil {
		h.logger.Error("创建或获取 Assistant 失败", elog.FieldErr(err))
		return err
	}

	// 3. 向 Thread 添加用户消息
	llmData := step.LLMData()
	h.logger.Info("📝 准备添加消息到 Thread",
		elog.String("threadID", ctx.Chat.LLMConversation.ID),
		elog.String("prompt", llmData.RenderedUserPrompt[:min(100, len(llmData.RenderedUserPrompt))]))

	_, err = h.client.Beta.Threads.Messages.New(
		ctx.Ctx,
		ctx.Chat.LLMConversation.ID, // Thread ID
		openai.BetaThreadMessageNewParams{
			Role: openai.BetaThreadMessageNewParamsRoleUser,
			Content: openai.BetaThreadMessageNewParamsContentUnion{
				OfString: openai.String(llmData.RenderedUserPrompt),
			},
		},
		h.options...,
	)
	if err != nil {
		h.logger.Error("❌ 添加消息到 Thread 失败", elog.FieldErr(err))
		return err
	}
	h.logger.Info("✅ 消息已添加到 Thread")

	// 4. 转换 Functions 为 Tools
	h.logger.Info("🔧 开始转换 Functions 为 Tools", elog.Int("functionCount", len(cfg.Functions)))
	tools := make([]openai.AssistantToolUnionParam, 0, len(cfg.Functions))
	for _, fn := range cfg.Functions {
		var funcDef shared.FunctionDefinitionParam
		err := json.Unmarshal([]byte(fn.Definition), &funcDef)
		if err != nil {
			h.logger.Error("❌ 函数定义反序列化失败", elog.String("function", fn.Definition), elog.FieldErr(err))
			continue
		}
		tools = append(tools, openai.AssistantToolParamOfFunction(funcDef))
		h.logger.Debug("✅ 已添加函数工具", elog.String("name", fn.Name))
	}
	h.logger.Info("✅ Functions 转换完成", elog.Int("toolCount", len(tools)))

	// 5. 创建并流式运行 Run（在 Run 级别传递完整配置）
	h.logger.Info("🚀 开始创建 Run",
		elog.String("threadID", ctx.Chat.LLMConversation.ID),
		elog.String("assistantID", assistantID),
		elog.Int("toolsCount", len(tools)),
		elog.String("systemPrompt", cfg.SystemPrompt[:min(200, len(cfg.SystemPrompt))]),
		elog.Any("temperature", cfg.Temperature),
		elog.Any("topP", cfg.TopP),
		elog.Int("maxTokens", cfg.MaxTokens))

	s := h.client.Beta.Threads.Runs.NewStreaming(
		ctx.Ctx,
		ctx.Chat.LLMConversation.ID, // Thread ID
		openai.BetaThreadRunNewParams{
			AssistantID:         assistantID,
			Instructions:        openai.String(cfg.SystemPrompt),        // 覆盖 Assistant 的 Instructions
			Temperature:         openai.Float(float64(cfg.Temperature)), // 覆盖 Temperature
			TopP:                openai.Float(float64(cfg.TopP)),        // 覆盖 TopP
			MaxCompletionTokens: openai.Int(int64(cfg.MaxTokens)),       // 设置最大 token 数
			Tools:               tools,                                  // 覆盖 Tools
			ToolChoice: openai.AssistantToolChoiceOptionUnionParam{
				OfAuto: openai.String("required"), // 强制调用工具
			},
		},
		h.options...,
	)

	// 6. 处理流式响应
	h.logger.Info("🔄 开始处理流式响应")
	err = h.processRunStream(ctx, cfg, s)
	if err != nil {
		h.logger.Error("❌ 处理流式响应失败", elog.FieldErr(err))
	} else {
		h.logger.Info("✅ Stream 方法执行完成")
	}
	return err
}

// initThreadIfNeeded 初始化 Thread 并把 Thread ID 放入到 ctx.Chat.LLMConversation.ID 中
func (h *Handler) initThreadIfNeeded(ctx *domain.StreamContext) error {
	if ctx.Chat.LLMConversation.ID != "" {
		return nil
	}
	thread, err := h.client.Beta.Threads.New(ctx.Ctx, openai.BetaThreadNewParams{}, h.options...)
	if err != nil {
		return err
	}
	ctx.Chat.LLMConversation.ID = thread.ID
	return nil
}

// getOrCreateAssistant 获取或创建 Assistant（每个 Chat.Sn 对应一个 Assistant）
// Assistant 作为配置模板，实际的 SystemPrompt、Tools 等会在 Run 中覆盖
func (h *Handler) getOrCreateAssistant(ctx *domain.StreamContext, cfg domain.InvocationConfigVersion) (string, error) {
	// 先从缓存查找
	if assistantID, ok := h.assistants.Load(ctx.Chat.Sn); ok {
		h.logger.Debug("使用缓存的 Assistant",
			elog.String("sn", ctx.Chat.Sn),
			elog.String("assistantID", assistantID.(string)))
		return assistantID.(string), nil
	}

	// 创建 Assistant（最小配置，作为模板）
	h.logger.Info("创建新的 Assistant",
		elog.String("sn", ctx.Chat.Sn),
		elog.String("model", cfg.Model.Name))

	assistant, err := h.client.Beta.Assistants.New(ctx.Ctx, openai.BetaAssistantNewParams{
		Model: cfg.Model.Name,                       // 设置默认模型
		Name:  openai.String("Chat-" + ctx.Chat.Sn), // 便于识别
		// Instructions、Tools 等在 Run 中动态设置，不在 Assistant 中固定
	}, h.options...)
	if err != nil {
		h.logger.Error("创建 Assistant 失败", elog.FieldErr(err))
		return "", err
	}

	// 缓存 Assistant ID
	h.assistants.Store(ctx.Chat.Sn, assistant.ID)
	h.logger.Info("Assistant 创建成功并已缓存",
		elog.String("assistantID", assistant.ID),
		elog.String("sn", ctx.Chat.Sn))

	return assistant.ID, nil
}

// processRunStream 处理 Run 的流式响应
func (h *Handler) processRunStream(
	ctx *domain.StreamContext,
	cfg domain.InvocationConfigVersion,
	stream *ssestream.Stream[openai.AssistantStreamEventUnion],
) error {
	var pendingToolCalls []openai.RequiredActionFunctionToolCall
	var runID string

	for stream.Next() {
		event := stream.Current()

		// 尝试获取事件的详细信息
		eventJSON, _ := json.Marshal(event)
		h.logger.Info("📨 收到事件",
			elog.String("event", event.Event),
			elog.String("eventDetail", string(eventJSON)[:min(500, len(string(eventJSON)))]))

		switch event.Event {
		case "thread.message.delta":
			// 文本消息增量（输出 warning 日志）
			msg := event.AsThreadMessageDelta()
			if len(msg.Data.Delta.Content) > 0 {
				for _, content := range msg.Data.Delta.Content {
					if text := content.AsText(); text.Text.Value != "" {
						h.logger.Warn("LLM输出非预期文本",
							elog.String("Delta", text.Text.Value))
					}
				}
			}

		case "thread.run.step.delta":
			// 工具调用步骤增量 - 记录日志
			h.logger.Info("收到 run step delta 事件")

		case "thread.run.step.completed":
			// 步骤完成，记录日志
			step := event.AsThreadRunStepCompleted()
			h.logger.Info("📦 run step completed", elog.String("stepID", step.Data.ID))

		case "thread.run.requires_action":
			// 需要执行 function calls
			run := event.AsThreadRunRequiresAction()
			runID = run.Data.ID
			pendingToolCalls = run.Data.RequiredAction.SubmitToolOutputs.ToolCalls
			h.logger.Info("🔧 Run 需要执行工具调用",
				elog.String("runID", runID),
				elog.Int("toolCallCount", len(pendingToolCalls)))

		case "thread.run.created":
			// Run 已创建
			h.logger.Debug("Run 已创建")

		case "thread.run.queued":
			// Run 已入队
			h.logger.Debug("Run 已入队")

		case "thread.run.in_progress":
			// Run 正在执行
			h.logger.Debug("Run 正在执行")

		case "thread.run.completed":
			h.logger.Info("✅ Run 完成")
			// 如果完成时没有调用任何工具，说明 LLM 可能直接返回了文本
			if len(pendingToolCalls) == 0 {
				h.logger.Warn("⚠️ Run 完成但没有调用任何工具，可能 ToolChoice 设置未生效")
			}

		case "thread.run.failed":
			failedRun := event.AsThreadRunFailed()
			errMsg := "Run 失败"
			if failedRun.Data.LastError.Message != "" {
				errMsg = failedRun.Data.LastError.Message
			}
			h.logger.Error("❌ Run 失败", elog.String("error", errMsg))
			return errors.New(errMsg)

		case "error":
			h.logger.Error("❌ 百炼返回错误事件")
			return errors.New("stream error event received")

		default:
			// 记录未处理的事件
			h.logger.Warn("⚠️ 未处理的事件类型", elog.String("event", event.Event))
		}
	}

	h.logger.Info("📊 Stream 处理完成",
		elog.Int("pendingToolCalls", len(pendingToolCalls)))

	// 处理待执行的 function calls
	if len(pendingToolCalls) > 0 {
		return h.handleFunctionCalls(ctx, cfg, runID, pendingToolCalls)
	}

	return stream.Err()
}

// handleFunctionCalls 处理 function calls 的执行和结果提交
func (h *Handler) handleFunctionCalls(
	ctx *domain.StreamContext,
	cfg domain.InvocationConfigVersion,
	runID string,
	toolCalls []openai.RequiredActionFunctionToolCall,
) error {
	toolOutputs := make([]openai.BetaThreadRunSubmitToolOutputsParamsToolOutput, 0, len(toolCalls))
	seenName := make(map[string]int)
	seenNextInvCfgID := make(map[string]int)
	nextInvCfgIDs := make([]int64, 0, len(toolCalls))

	// 执行每个 function call
	for _, toolCall := range toolCalls {
		h.logger.Info("🔧 开始处理 function call",
			elog.String("id", toolCall.ID),
			elog.String("function", toolCall.Function.Name),
			elog.String("arguments", toolCall.Function.Arguments))

		// 处理重复的函数调用
		if idx, ok := seenName[toolCall.Function.Name]; ok {
			p := toolOutputs[idx]
			p.ToolCallID = openai.String(toolCall.ID)
			toolOutputs = append(toolOutputs, p)
			if idx2, ok := seenNextInvCfgID[toolCall.Function.Name]; ok {
				nextInvCfgIDs = append(nextInvCfgIDs, nextInvCfgIDs[idx2])
			}
			continue
		}

		// 从 registry 查找并执行
		fn, err := h.registry.Lookup(toolCall.Function.Name)
		if err != nil {
			h.logger.Error("❌ 函数调用未找到",
				elog.String("函数名", toolCall.Function.Name),
				elog.FieldErr(err))
			return err
		}

		h.logger.Debug("✅ 找到 function，开始执行", elog.String("function", toolCall.Function.Name))
		fcallResp, err := fn.Call(ctx, fcall.Request{Args: []byte(toolCall.Function.Arguments)})
		if err != nil {
			h.logger.Error("❌ 执行函数调用失败",
				elog.String("函数名", toolCall.Function.Name),
				elog.String("参数值", toolCall.Function.Arguments),
				elog.FieldErr(err))
		} else {
			h.logger.Info("✅ Function 执行成功",
				elog.String("function", toolCall.Function.Name),
				elog.String("content", fcallResp.Content))

			if fcallResp.NextInvCfgID > 0 {
				nextInvCfgIDs = append(nextInvCfgIDs, fcallResp.NextInvCfgID)
				seenNextInvCfgID[toolCall.Function.Name] = len(nextInvCfgIDs) - 1
			}
		}

		// 构建工具输出
		h.logger.Debug("📤 准备返回 function call 结果给百炼",
			elog.String("toolCallID", toolCall.ID),
			elog.String("content", fcallResp.Content))

		toolOutputs = append(toolOutputs, openai.BetaThreadRunSubmitToolOutputsParamsToolOutput{
			ToolCallID: openai.String(toolCall.ID),
			Output:     openai.String(fcallResp.Content),
		})
		seenName[toolCall.Function.Name] = len(toolOutputs) - 1
	}

	h.logger.Info("📤 收集到的 Function Call Outputs",
		elog.Int("count", len(toolOutputs)),
		elog.Any("nextInvCfgIDs", nextInvCfgIDs))

	// 提交工具输出并流式获取后续响应
	h.logger.Info("🔄 使用流式 API 提交函数结果并获取 LLM 响应",
		elog.String("runID", runID),
		elog.String("threadID", ctx.Chat.LLMConversation.ID))

	stream := h.client.Beta.Threads.Runs.SubmitToolOutputsStreaming(
		ctx.Ctx,
		ctx.Chat.LLMConversation.ID, // Thread ID
		runID,
		openai.BetaThreadRunSubmitToolOutputsParams{
			ToolOutputs: toolOutputs,
		},
		h.options...,
	)

	// 递归处理流式响应
	err := h.processRunStream(ctx, cfg, stream)
	if err != nil {
		h.logger.Error("❌ 处理函数调用后的响应失败", elog.FieldErr(err))
		return err
	}

	// 处理需要执行的下一个配置
	for _, id := range nextInvCfgIDs {
		turn := ctx.Chat.LastTurn()
		turn.AssistantRun.StartLLMStep(id)
		h.logger.Warn("🔄 需要执行下一个 LLM 调用", elog.Int64("nextCfgID", id))
		return h.Handler.Stream(ctx)
	}

	return nil
}
