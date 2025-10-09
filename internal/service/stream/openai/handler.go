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
	"log"

	ai "github.com/ecodeclub/ai-gateway-go/api/proto/gen/chat/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/packages/ssestream"
	"github.com/openai/openai-go/v2/responses"
)

type Handler struct {
	client   openai.Client
	logger   *elog.Component
	registry *fcall.Registry
}

func NewHandler(
	client openai.Client,
	registry *fcall.Registry,
) *Handler {
	return &Handler{
		client:   client,
		registry: registry,
		logger:   elog.DefaultLogger.With(elog.FieldComponentName("openai.Handler")),
	}
}

func (h *Handler) Stream(ctx *domain.StreamContext) error {
	params := h.newParams(ctx)
	val, _ := json.Marshal(params)
	h.logger.Debug(string(val))
	stream := h.client.Responses.NewStreaming(ctx.Ctx, params)
	return h.forward(ctx, stream)
}

func (h *Handler) newParams(ctx *domain.StreamContext) responses.ResponseNewParams {
	input := h.toInput(ctx)
	h.logger.Debug("调用 OpenAI 的输入", elog.Any("input", input))
	step := ctx.Chat.LastTurn().AssistantRun.LastStep()
	cfg := step.LLMData().Cfg
	params := responses.ResponseNewParams{
		Input: input,
		Model: cfg.Model.Name,
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
	for _, src := range history {
		if src.Content == "" {
			continue
		}
		items = append(items, h.toInputItem(src.Content, src.Role))
	}
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

func (h *Handler) forward(ctx *domain.StreamContext, stream *ssestream.Stream[responses.ResponseStreamEventUnion]) error {
	// defer close(events)
	var responseID string
	for stream.Next() {
		event := stream.Current()
		switch event.Type {
		case "error":
			h.sendEvt(ctx, domain.StreamEventV1{
				Err: errors.New(event.AsError().Message),
			})
		case "response.created":
			responseID = event.Response.ID
			log.Printf("forward: response created ID = %s\n", responseID)
		case "response.output_text.delta":
			text := event.AsResponseOutputTextDelta()
			h.sendEvt(ctx, domain.StreamEventV1{
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
			h.logger.Debug("收到 function call 调用请求", elog.String("function", fc.Name))
			fn, err := h.registry.Lookup(fc.Name)
			if err != nil {
				h.logger.Error("函数调用未找到",
					elog.FieldCustomKeyValue("函数名", fc.Name),
					elog.FieldErr(err),
				)
				return err
			}

			_, err = fn.Call(ctx, fcall.Request{Args: []byte(fc.Arguments)})
			if err != nil {
				h.logger.Error("执行函数调用失败",
					elog.FieldCustomKeyValue("函数名", fc.Name),
					elog.FieldCustomKeyValue("参数值", fc.Arguments),
					elog.FieldErr(err),
				)
				return err
			}
		}
	}

	return stream.Err()
}

func (h *Handler) sendEvt(ctx *domain.StreamContext, evt domain.StreamEventV1) {
	err := ctx.Sender.Send(evt)
	if err != nil {
		h.logger.Error("发送数据失败", elog.FieldErr(err))
	}
}
