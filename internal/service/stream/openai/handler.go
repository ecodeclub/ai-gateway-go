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

package openai

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	ai "github.com/ecodeclub/ai-gateway-go/api/proto/gen/chat/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/conversations"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/responses"
)

type Handler struct {
	client   openai.Client
	logger   *elog.Component
	registry *fcall.Registry
	// 发送给 OpenAI 需要携带的一些公共的头部
	options []option.RequestOption

	// 发起下一个 LLM 调用
	Handler stream.Handler
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
		client:   client,
		registry: registry,
		options:  headerOpts,
		logger:   elog.DefaultLogger.With(elog.FieldComponentName("openai.Handler")),
	}
}

func (h *Handler) Stream(ctx *domain.StreamContext) error {
	err := h.initConversationsIfNeeded(ctx)
	if err != nil {
		return err
	}
	step := ctx.Chat.LastTurn().AssistantRun.LastStep()
	cfg := step.LLMData().Cfg
	params := h.newParams(ctx, cfg)
	// 这种比较复杂的打印日志的代码，就用一个判断来减少线上消耗
	if elog.DebugLevel == elog.DebugLevel {
		val, _ := json.Marshal(params)
		h.logger.Debug(string(val))
	}
	s := h.client.Responses.NewStreaming(ctx.Ctx, params, h.options...)
	return h.forward(ctx, cfg, s)
}

// initConversationsIfNeeded 初始化并且把 cid3rd 放入到 ctx.Chat 里面
func (h *Handler) initConversationsIfNeeded(ctx *domain.StreamContext) error {
	if ctx.Chat.LLMConversation.ID != "" {
		return nil
	}
	c, err := h.client.Conversations.New(ctx.Ctx, conversations.ConversationNewParams{}, h.options...)
	if err != nil {
		return err
	}
	ctx.Chat.LLMConversation.ID = c.ID
	return nil
}

func (h *Handler) newParams(ctx *domain.StreamContext, cfg domain.InvocationConfigVersion) responses.ResponseNewParams {
	input := h.toInput(ctx)
	h.logger.Debug("调用 OpenAI 的输入", elog.String("cid3rd", ctx.LLMCid()), elog.Any("input", input))
	params := responses.ResponseNewParams{
		Input: input,
		Model: cfg.Model.Name,
		Conversation: responses.ResponseNewParamsConversationUnion{
			OfConversationObject: &responses.ResponseConversationParam{
				ID: ctx.Chat.LLMConversation.ID,
			},
		},
		Tools: slice.Map(cfg.Functions, func(_ int, src domain.Function) responses.ToolUnionParam {
			var p responses.FunctionToolParam
			err := json.Unmarshal([]byte(src.Definition), &p)
			if err != nil {
				h.logger.Error("函数定义反序列化失败", elog.String("function", src.Definition), elog.FieldErr(err))
			}
			return responses.ToolUnionParam{OfFunction: &p}
		}),
		Temperature:     openai.Float(float64(cfg.Temperature)),
		TopP:            openai.Float(float64(cfg.TopP)),
		MaxOutputTokens: openai.Int(int64(cfg.MaxTokens)),
	}
	if cfg.SystemPrompt != "" {
		params.Instructions = openai.String(cfg.SystemPrompt)
	}
	return params
}

func (h *Handler) toInput(ctx *domain.StreamContext) responses.ResponseNewParamsInputUnion {
	history := ctx.History()
	items := make([]responses.ResponseInputItemUnionParam, 0, len(history)+1)
	// 把这一次输入加入进去，这里固定将 role 设置为 user，因为这里有一个隐含假设，这一个输入一定是 RoleUser
	// 只能用当前 step
	llmData := ctx.Chat.LastTurn().AssistantRun.LastStep().LLMData()
	items = append(items, h.toInputItem(llmData.RenderedUserPrompt, ai.RoleUser))
	return responses.ResponseNewParamsInputUnion{
		OfInputItemList: items,
	}
}

func (h *Handler) toInputItem(content, role string) responses.ResponseInputItemUnionParam {
	switch role {
	case ai.RoleUser:
		contentList := responses.ResponseInputMessageContentListParam{
			{
				OfInputText: &responses.ResponseInputTextParam{Text: content},
			},
		}
		return responses.ResponseInputItemUnionParam{
			OfInputMessage: &responses.ResponseInputItemMessageParam{
				Content: contentList,
				Role:    ai.RoleUser,
			},
		}
	case ai.RoleAssistant:
		contentList := []responses.ResponseOutputMessageContentUnionParam{
			{
				OfOutputText: &responses.ResponseOutputTextParam{Text: content},
			},
		}
		return responses.ResponseInputItemUnionParam{
			OfOutputMessage: &responses.ResponseOutputMessageParam{
				Content: contentList,
				Role:    ai.RoleAssistant,
			},
		}
	default:
		h.logger.Error("未知 role", elog.String("role", role))
		return responses.ResponseInputItemUnionParam{}
	}
}

func (h *Handler) forward(ctx *domain.StreamContext,
	cfg domain.InvocationConfigVersion,
	stream *ssestream.Stream[responses.ResponseStreamEventUnion]) error {
	// defer close(events)
	textDeltaCount := 0
	//functionCallCount := 0
	var pendingFunctionCalls []responses.ResponseFunctionToolCall // 收集所有 function calls

	for stream.Next() {
		event := stream.Current()
		h.logger.Debug("📨 收到事件", elog.String("type", event.Type))

		switch event.Type {
		case "error":
			h.logger.Error("❌ OpenAI 返回错误", elog.String("error", event.AsError().Message))
			h.sendEvt(ctx, domain.StreamEventV1{
				Err: errors.New(event.AsError().Message),
			})
		case "response.output_text.delta":
			textDeltaCount++
			text := event.AsResponseOutputTextDelta()
			h.sendEvt(ctx, domain.StreamEventV1{
				Delta: &domain.Delta{
					Content: text.Delta,
				},
			})
		case "response.output_item.done":
			item := event.AsResponseOutputItemDone().Item
			h.logger.Debug("📦 output_item.done", elog.String("itemType", item.Type))
			if item.Type != "function_call" {
				h.logger.Debug("非 function call 类型的 output item", elog.String("type", item.Type))
				continue
			}

			fc := item.AsFunctionCall()
			h.logger.Warn("🔧 LLM 调用了 function call",
				elog.Any("fc", fc),
				elog.String("ID", fc.ID),
				elog.String("CallID", fc.CallID),
				elog.String("Name", fc.Name),
				elog.String("Arguments", fc.Arguments),
				elog.Int("textDeltaCount", textDeltaCount))

			// 不立即处理，先收集起来
			pendingFunctionCalls = append(pendingFunctionCalls, fc)
			//functionCallCount = len(pendingFunctionCalls)
		}
	}

	h.logger.Info("✅ Stream 处理完成",
		elog.Int("textDeltaCount", textDeltaCount),
		elog.Int("functionCallCount", len(pendingFunctionCalls)))

	// Stream 完成后，再处理 function calls
	if len(pendingFunctionCalls) > 0 {
		h.logger.Info("🔄 开始处理待处理的 function calls", elog.Int("count", len(pendingFunctionCalls)))
		err := h.handleFC(ctx, cfg, pendingFunctionCalls)
		if err != nil {
			return err
		}
	}

	return stream.Err()
}

func (h *Handler) handleFC(ctx *domain.StreamContext, cfg domain.InvocationConfigVersion, fcs []responses.ResponseFunctionToolCall) error {
	params := make([]responses.ResponseInputItemFunctionCallOutputParam, 0, len(fcs))
	seenName := make(map[string]int)
	seenNextInvCfgID := make(map[string]int)
	nextInvCfgIDs := make([]int64, 0, len(fcs))
	for _, fc := range fcs {
		h.logger.Info("🔧 开始处理 function call",
			elog.String("function", fc.Name),
			elog.String("callID", fc.CallID),
			elog.String("arguments", fc.Arguments))

		if idx, ok := seenName[fc.Name]; ok {
			p := params[idx]
			p.CallID = fc.CallID
			params = append(params, p)
			if idx2, ok := seenNextInvCfgID[fc.Name]; ok {
				nextInvCfgIDs = append(nextInvCfgIDs, nextInvCfgIDs[idx2])
			}
			continue
		}

		fn, err := h.registry.Lookup(fc.Name)
		if err != nil {
			h.logger.Error("❌ 函数调用未找到",
				elog.String("函数名", fc.Name),
				elog.FieldErr(err),
			)
			return err
		}

		h.logger.Debug("✅ 找到 function，开始执行", elog.String("function", fc.Name))
		fcallResp, err := fn.Call(ctx, fcall.Request{Args: []byte(fc.Arguments)})
		if err != nil {
			h.logger.Error("❌ 执行函数调用失败",
				elog.String("函数名", fc.Name),
				elog.String("参数值", fc.Arguments),
				elog.FieldErr(err),
			)
		} else {
			h.logger.Info("✅ Function 执行成功",
				elog.String("function", fc.Name),
				elog.String("content", fcallResp.Content))

			if fcallResp.NextInvCfgID > 0 {
				nextInvCfgIDs = append(nextInvCfgIDs, fcallResp.NextInvCfgID)
				seenNextInvCfgID[fc.Name] = len(nextInvCfgIDs) - 1
			}
		}

		// 不管有没有问题，都要返回一个 response
		h.logger.Debug("📤 准备返回 function call 结果给 OpenAI",
			elog.String("callID", fc.CallID),
			elog.String("content", fcallResp.Content))

		params = append(params, h.toFCResponseOfFunctionCallOutput(fc, fcallResp))
		seenName[fc.Name] = len(params) - 1
	}

	h.logger.Info("📤 收集到的Function Call Output",
		elog.Any("params", params),
		elog.Any("nextInvCfgIDs", nextInvCfgIDs))

	respBody := responses.ResponseNewParams{
		Model: cfg.Model.Name,
		Conversation: responses.ResponseNewParamsConversationUnion{
			OfConversationObject: &responses.ResponseConversationParam{
				ID: ctx.Chat.LLMConversation.ID,
			},
		},
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: slice.Map(params, func(_ int, src responses.ResponseInputItemFunctionCallOutputParam) responses.ResponseInputItemUnionParam {
				return responses.ResponseInputItemUnionParam{
					OfFunctionCallOutput: &src,
				}
			}),
		},
	}

	// 重试机制：处理 conversation lock 失败
	var resp *responses.Response
	var err1 error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// 【重要】：必须在一次响应将【所有】的Function Call Output传递回去，包含大模型重试的Function Call Output。
		resp, err1 = h.client.Responses.New(ctx.Ctx, respBody, h.options...)
		if err1 == nil {
			break
		}

		// 检查是否是 conversation lock 错误
		errMsg := err1.Error()
		if strings.Contains(errMsg, "conversation_lock_failed") || strings.Contains(errMsg, "Failed to acquire conversation lock") {
			h.logger.Warn("⚠️ Conversation lock 失败，等待重试",
				elog.Int("attempt", i+1),
				elog.Int("maxRetries", maxRetries),
				elog.Any("respBody", respBody))
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond) // 递增等待时间
			continue
		}

		// 其他错误，不重试
		break
	}

	if err1 != nil {
		h.logger.Error("❌ 返回 FC 响应给 OpenAI 失败（已重试）",
			elog.Any("respBody", respBody),
			elog.Int("retries", maxRetries),
			elog.FieldErr(err1))
		return err1
	}

	h.logger.Info("✅ 成功返回 function call 结果给 OpenAI",
		elog.Any("respBody", respBody),
		elog.String("responseID", resp.ID))

	for _, id := range nextInvCfgIDs {
		turn := ctx.Chat.LastTurn()
		turn.AssistantRun.StartLLMStep(id)
		h.logger.Warn("🔄 需要执行下一个 LLM 调用", elog.Int64("nextCfgID", id))
		return h.Handler.Stream(ctx)
	}
	return nil
}

func (h *Handler) toFCResponseOfFunctionCallOutput(fc responses.ResponseFunctionToolCall, fcallResp fcall.Response) responses.ResponseInputItemFunctionCallOutputParam {
	// 暂时固定为 completed
	const fcStatusCompleted = "completed"
	return responses.ResponseInputItemFunctionCallOutputParam{
		CallID: fc.CallID,
		Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
			OfString: param.NewOpt(fcallResp.Content),
		},
		Status: fcStatusCompleted,
	}
}

func (h *Handler) sendEvt(ctx *domain.StreamContext, evt domain.StreamEventV1) {
	err := ctx.Sender.Send(evt)
	if err != nil {
		h.logger.Error("发送数据失败", elog.FieldErr(err))
	}
}
