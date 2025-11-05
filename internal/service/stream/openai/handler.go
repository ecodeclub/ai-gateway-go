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
	"fmt"

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

func (h *Handler) Stream(ctx *domain.StreamContext) (stream.Response, error) {
	step := ctx.Chat.CurrentStep()
	cfg := step.Cfg
	params, err := h.newParams(ctx, cfg)
	if err != nil {
		return stream.Response{}, err
	}
	// 这种比较复杂的打印日志的代码，就用一个判断来减少线上消耗
	if h.logger.IsDebugMode() {
		val, _ := json.Marshal(params)
		h.logger.Debug(string(val))
	}
	s := h.client.Responses.NewStreaming(ctx.Ctx, params, h.options...)
	return h.forward(ctx, cfg, s)
}

// initConversationsIfNeeded 初始化并且把 cid3rd 放入到 ctx.Chat 里面
func (h *Handler) initConversationsIfNeeded(ctx *domain.StreamContext) error {
	step := ctx.Chat.CurrentStep()
	if step.Thread.Conversation.ID != "" {
		return nil
	}
	c, err := h.client.Conversations.New(ctx.Ctx, conversations.ConversationNewParams{}, h.options...)
	if err != nil {
		return fmt.Errorf("new conversation: %w", err)
	}
	step.Thread.Conversation.ID = c.ID
	return nil
}

func (h *Handler) newParams(ctx *domain.StreamContext, cfg domain.InvocationConfigVersion) (responses.ResponseNewParams, error) {
	input := h.toInput(ctx)
	step := ctx.Chat.CurrentTurn().AssistantRun.CurrentStep()
	h.logger.Debug("调用 OpenAI 的输入", elog.String("cid3rd", step.Thread.Conversation.ID), elog.Any("input", input))
	params := responses.ResponseNewParams{
		Input:           input,
		Model:           cfg.Model.Name,
		MaxOutputTokens: openai.Int(int64(cfg.MaxTokens)),
	}

	if step.Thread.Conversation.ID == "" {
		h.logger.Debug("初始化 conversation")
		err := h.initConversationsIfNeeded(ctx)
		if err != nil {
			return responses.ResponseNewParams{}, err
		}
		params.Conversation = responses.ResponseNewParamsConversationUnion{
			OfConversationObject: &responses.ResponseConversationParam{
				ID: step.Thread.Conversation.ID,
			},
		}
		// 设置 instructions
		params.Instructions = openai.String(cfg.SystemPrompt)
	}

	if len(cfg.Functions) > 0 {
		params.ToolChoice = responses.ResponseNewParamsToolChoiceUnion{
			OfToolChoiceMode: param.NewOpt(responses.ToolChoiceOptions("required")),
		}
		params.Tools = slice.Map(cfg.Functions, func(_ int, src domain.Function) responses.ToolUnionParam {
			var p responses.FunctionToolParam
			err := json.Unmarshal([]byte(src.Definition), &p)
			if err != nil {
				h.logger.Error("函数定义反序列化失败", elog.String("function", src.Definition), elog.FieldErr(err))
			}
			return responses.ToolUnionParam{OfFunction: &p}
		})
	}

	if cfg.Temperature >= 0 {
		params.Temperature = openai.Float(float64(cfg.Temperature))
	}

	if cfg.TopP >= 0 {
		params.TopP = openai.Float(float64(cfg.TopP))
	}
	return params, nil
}

func (h *Handler) toInput(ctx *domain.StreamContext) responses.ResponseNewParamsInputUnion {
	history := ctx.History()
	items := make([]responses.ResponseInputItemUnionParam, 0, len(history)+1)
	step := ctx.Chat.CurrentStep()
	items = append(items, h.toInputItem(step.RenderedUserPrompt, ai.RoleUser))
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
	sse *ssestream.Stream[responses.ResponseStreamEventUnion]) (stream.Response, error) {
	// 一般都只有一个
	fcallsItems := make([]responses.ResponseFunctionToolCall, 0, 1)
	for sse.Next() {
		event := sse.Current()
		switch event.Type {
		case "error":
			h.sendEvt(ctx, domain.StreamEvent{
				Err: errors.New(event.AsError().Message),
			})
		case "response.output_text.delta":
			text := event.AsResponseOutputTextDelta()
			h.sendEvt(ctx, domain.StreamEvent{
				Delta: &domain.Delta{
					Content: text.Delta,
				},
			})
		case "response.output_item.done":
			item := event.AsResponseOutputItemDone().Item
			if item.Type != "function_call" {
				h.logger.Debug("非 function call 类型的 output item", elog.String("type", item.Type))
				continue
			}
			fc := item.AsFunctionCall()
			fcallsItems = append(fcallsItems, fc)
		}
	}

	err := sse.Err()
	if len(fcallsItems) == 0 {
		return stream.Response{}, err
	}
	if err != nil {
		h.logger.Error("sse 有 ERROR", elog.FieldErr(err))
	}
	fcallRespList := make([]FCResp, 0, len(fcallsItems))
	for _, fc := range fcallsItems {
		fcallResp, err1 := h.handleFC(ctx, fc)
		if err1 != nil {
			h.logger.Error("执行 function call 出现问题", elog.FieldErr(err), elog.Any("fc", fc))
			continue
		}
		fcallRespList = append(fcallRespList, FCResp{
			FC:   fc,
			Resp: fcallResp,
		})
	}

	// 不管有没有问题，都要返回一个 response，一次性返回所有的 function call 的结果
	fcRespInput := h.toFCResulInput(ctx, cfg, fcallRespList)
	if h.logger.IsDebugMode() {
		val, _ := json.Marshal(fcRespInput)
		h.logger.Debug(string(val))
	}
	_, err1 := h.client.Responses.New(ctx.Ctx, fcRespInput, h.options...)
	if err1 != nil {
		h.logger.Error("返回 FC 响应给 OpenAI 失败",
			elog.FieldErr(err1))
	}
	nextState := ""
	for _, fc := range fcallRespList {
		fcallResp := fc.Resp
		if fcallResp.NextState != "" {
			nextState = fcallResp.NextState
		}
	}
	return stream.Response{NextState: nextState}, err
}

func (h *Handler) handleFC(ctx *domain.StreamContext, fc responses.ResponseFunctionToolCall) (fcall.Response, error) {
	h.logger.Debug("收到 function call 调用请求", elog.String("function", fc.Name))
	fn, err := h.registry.Lookup(fc.Name)
	if err != nil {
		h.logger.Error("函数调用未找到",
			elog.FieldCustomKeyValue("函数名", fc.Name),
			elog.FieldErr(err),
		)
		return fcall.Response{}, err
	}
	return fn.Call(ctx, fcall.Request{Args: []byte(fc.Arguments)})
}

func (h *Handler) toFCResulInput(ctx *domain.StreamContext, cfg domain.InvocationConfigVersion, respList []FCResp) responses.ResponseNewParams {
	// 暂时固定为 completed
	const fcStatusCompleted = "completed"
	step := ctx.Chat.CurrentStep()
	return responses.ResponseNewParams{
		Model: cfg.Model.Name,
		Conversation: responses.ResponseNewParamsConversationUnion{
			OfConversationObject: &responses.ResponseConversationParam{
				ID: step.Thread.Conversation.ID,
			},
		},
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: slice.Map[FCResp, responses.ResponseInputItemUnionParam](respList, func(idx int, src FCResp) responses.ResponseInputItemUnionParam {
				return responses.ResponseInputItemUnionParam{
					OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
						CallID: src.FC.CallID,
						Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
							OfString: param.NewOpt(src.Resp.Content),
						},
						Status: fcStatusCompleted,
					},
				}
			}),
		},
	}
}

func (h *Handler) sendEvt(ctx *domain.StreamContext, evt domain.StreamEvent) {
	err := ctx.Sender.Send(evt)
	if err != nil {
		h.logger.Error("发送数据失败", elog.FieldErr(err))
	}
}

type FCResp struct {
	FC   responses.ResponseFunctionToolCall
	Resp fcall.Response
}
