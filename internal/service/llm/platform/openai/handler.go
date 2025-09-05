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
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/packages/ssestream"
	"github.com/openai/openai-go/v2/responses"
)

type Handler struct {
	client           openai.Client
	funcCallRegistry *fcall.Registry
	logger           *elog.Component
}

func NewHandler(
	client openai.Client,
	funcCallRegistry *fcall.Registry,
) *Handler {
	return &Handler{
		client:           client,
		funcCallRegistry: funcCallRegistry,
		logger:           elog.DefaultLogger.With(elog.FieldComponentName("openai.Handler")),
	}
}

func (h *Handler) Stream(ctx context.Context, req domain.StreamRequest) (chan domain.StreamEvent, error) {
	log.Printf("stream request: %#v\n", req)
	events := make(chan domain.StreamEvent, 10)
	go func() {
		h.forward(req, h.stream(ctx, req), events)
	}()
	return events, nil
}

func (h *Handler) stream(ctx context.Context, req domain.StreamRequest) (stream *ssestream.Stream[responses.ResponseStreamEventUnion]) {
	return h.client.Responses.NewStreaming(ctx, h.newParams(req))
}

func (h *Handler) newParams(req domain.StreamRequest) responses.ResponseNewParams {
	params := responses.ResponseNewParams{
		Input: h.toInput(req.Messages),
		Model: req.ConfigVersion.Model.Name,
		Tools: slice.Map(req.ConfigVersion.Functions, func(_ int, src domain.Function) responses.ToolUnionParam {
			var p responses.FunctionToolParam
			_ = json.Unmarshal([]byte(src.Definition), &p)
			return responses.ToolUnionParam{OfFunction: &p}
		}),
		Temperature:     openai.Float(float64(req.ConfigVersion.Temperature)),
		TopP:            openai.Float(float64(req.ConfigVersion.TopP)),
		MaxOutputTokens: openai.Int(int64(req.ConfigVersion.MaxTokens)),
		// PreviousResponseID: openai.String(req.PreviousResponseID),
		// Store:              openai.Bool(true),
		// Conversation: responses.ResponseNewParamsConversationUnion{
		// 	OfConversationObject: &responses.ResponseConversationParam{
		// 		ID: req.ConversationID,
		// 	},
		// },
	}
	// if req.CallID != "" {
	// 	params.Input.OfInputItemList = append(params.Input.OfInputItemList, responses.ResponseInputItemUnionParam{
	// 		OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
	// 			CallID: req.CallID,
	// 			Output: "success",
	// 			Status: "completed",
	// 		},
	// 	})
	// }
	if req.ConfigVersion.SystemPrompt != "" {
		params.Instructions = openai.String(req.ConfigVersion.SystemPrompt)
	}
	return params
}

func (h *Handler) toInput(messages []domain.Message) responses.ResponseNewParamsInputUnion {
	return responses.ResponseNewParamsInputUnion{
		OfInputItemList: slice.Map(messages, func(idx int, src domain.Message) responses.ResponseInputItemUnionParam {
			contentList := responses.ResponseInputMessageContentListParam{
				{
					OfInputText: &responses.ResponseInputTextParam{Text: src.Content},
				},
			}
			switch src.Role {
			case domain.USER:
				return responses.ResponseInputItemUnionParam{
					OfInputMessage: &responses.ResponseInputItemMessageParam{
						Content: contentList,
						Role:    "user",
					},
				}
			case domain.SYSTEM:
				return responses.ResponseInputItemUnionParam{
					OfInputMessage: &responses.ResponseInputItemMessageParam{
						Content: contentList,
						Role:    "system",
					},
				}
			default:
				return responses.ResponseInputItemUnionParam{}
			}
		}),
	}
}

func (h *Handler) forward(req domain.StreamRequest, stream *ssestream.Stream[responses.ResponseStreamEventUnion], events chan domain.StreamEvent) {
	// defer close(events)
	for i := range req.Messages {
		log.Printf("req.messages[%d]: %#v\n", i, req.Messages[i])
	}
	fCtx := &fcall.Context{Context: context.Background()}
	var hasLLMInvokeCall bool
	var responseID string
	for stream.Next() {
		event := stream.Current()
		switch event.Type {
		case "error":
			events <- domain.StreamEvent{
				Error: errors.New(event.AsError().Message),
				Done:  true,
			}
		case "response.created":
			responseID = event.Response.ID
			log.Printf("forward: response created ID = %s\n", responseID)
		case "response.completed":
			if !hasLLMInvokeCall {
				events <- domain.StreamEvent{
					Done: true,
				}
			}
		case "response.output_text.delta":
			text := event.AsResponseOutputTextDelta()
			events <- domain.StreamEvent{
				Content: text.Delta,
			}
		case "response.reasoning_summary_text.delta":
			r := event.AsResponseReasoningSummaryTextDelta()
			events <- domain.StreamEvent{
				ReasoningContent: r.Delta,
			}
		case "response.output_item.done":
			item := event.AsResponseOutputItemDone().Item
			if item.Type != "function_call" {
				continue
			}

			fc := item.AsFunctionCall()
			fn, err := h.funcCallRegistry.Lookup(fc.Name)
			if err != nil {
				h.logger.Error("函数调用未找到",
					elog.FieldCustomKeyValue("函数名", fc.Name),
					elog.FieldErr(err),
				)
				// TODO 没找到返回错误？
				continue
			}

			_, err = fn.Call(fCtx, fcall.Request{Args: []byte(fc.Arguments)})
			if err != nil {
				h.logger.Error("执行函数调用失败",
					elog.FieldCustomKeyValue("函数名", fc.Name),
					elog.FieldCustomKeyValue("参数值", fc.Arguments),
					elog.FieldErr(err),
				)
			}
			switch fn.Name() {
			case fcall.NameAskUser:
				events <- domain.StreamEvent{
					Content: fCtx.GetAttachment(fcall.NameAskUser),
				}
			case fcall.NameEmitJSON:
				log.Printf("emit_json: 提取到的结构化数据: %#v\n", fCtx.JSONData)
			case fcall.NameInvokeLLM:
				hasLLMInvokeCall = true
				var msg domain.Message
				_ = json.Unmarshal([]byte(fCtx.GetAttachment(fcall.NameInvokeLLM)), &msg)
				log.Printf("invok_llm: %#v", msg)
				req.Messages = append(req.Messages, msg)
				params := h.newParams(req)
				// params.PreviousResponseID = openai.String(responseID)
				// params.Input.OfInputItemList = append(params.Input.OfInputItemList, responses.ResponseInputItemUnionParam{
				// 	OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
				// 		CallID: fc.CallID,
				// 		Output: msg.Content,
				// 		Status: "completed",
				// 	},
				// })
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
				h.forward(req, h.client.Responses.NewStreaming(ctx, params), events)
				cancel()
			}
		}
	}

	err := stream.Err()
	if err != nil {
		h.logger.Error("forward过程中出错", elog.FieldErr(err))
		log.Printf("forward过程中出错: %+v\n", err)
		// events <- domain.StreamEvent{
		// 	Error: err,
		// 	Done:  true,
		// }
	}
}
