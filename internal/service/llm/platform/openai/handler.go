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
	// if req.ConversationID == "" {
	// 	res, err := h.client.Conversations.New(ctx, conversations.ConversationNewParams{})
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	req.ConversationID = res.ID
	//
	// }
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
	messages := h.getMessages(req)
	log.Printf("messages: %#v\n", messages)
	params := responses.ResponseNewParams{
		Input: h.toInput(messages),
		Model: req.ConfigVersion.Model.Name,
		Tools: slice.Map(req.ConfigVersion.Functions, func(_ int, src domain.Function) responses.ToolUnionParam {
			var p responses.FunctionToolParam
			_ = json.Unmarshal([]byte(src.Definition), &p)
			return responses.ToolUnionParam{OfFunction: &p}
		}),
		Temperature:        openai.Float(float64(req.ConfigVersion.Temperature)),
		TopP:               openai.Float(float64(req.ConfigVersion.TopP)),
		MaxOutputTokens:    openai.Int(int64(req.ConfigVersion.MaxTokens)),
		PreviousResponseID: openai.String(req.PreviousResponseID),
		Store:              openai.Bool(true),
		// Conversation: responses.ResponseNewParamsConversationUnion{
		// 	OfConversationObject: &responses.ResponseConversationParam{
		// 		ID: req.ConversationID,
		// 	},
		// },
	}
	if req.CallID != "" {
		params.Input.OfInputItemList = append(params.Input.OfInputItemList, responses.ResponseInputItemUnionParam{
			OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
				CallID: req.CallID,
				Output: "success",
				Status: "completed",
			},
		})
	}
	return params
}

func (h *Handler) getMessages(req domain.StreamRequest) []domain.Message {
	systemContent := ""
	userContent := ""
	if req.ConfigVersion.SystemPrompt != "" {
		systemContent += req.ConfigVersion.SystemPrompt
		systemContent += "\n"
	}
	if req.ConfigVersion.Prompt != "" {
		userContent += req.ConfigVersion.Prompt
		userContent += "\n"
	}
	for _, message := range req.Messages {
		if message.Role == domain.SYSTEM && message.Content != "" {
			systemContent += message.Content
			systemContent += "\n"
		}
		if message.Role == domain.USER && message.Content != "" {
			userContent += message.Content
			userContent += "\n"
		}
	}
	messages := make([]domain.Message, 0, 2)
	if systemContent != "" {
		messages = append(messages, domain.Message{
			Role:    domain.SYSTEM,
			Content: systemContent,
		})
	}
	if userContent != "" {
		messages = append(messages, domain.Message{
			Role:    domain.USER,
			Content: userContent,
		})
	}
	return messages
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
	fCtx := &fcall.Context{Context: context.Background()}
	funcCallOutputs := make([]responses.ResponseInputItemUnionParam, 0, 10)
	var responseID string
	for stream.Next() {

		event := stream.Current()
		switch event.Type {
		case "error":
			events <- domain.StreamEvent{
				ResponseID: responseID,
				Error:      errors.New(event.AsError().Message),
				Done:       true,
			}
		case "response.created":
			responseID = event.Response.ID
			log.Printf("forward: response created ID = %s\n", responseID)
		case "response.completed":
			noFuncCalls := len(funcCallOutputs) == 0
			if noFuncCalls {
				evt := domain.StreamEvent{
					ResponseID: responseID,
					Done:       true,
				}
				events <- evt
				log.Printf("没有函数调用需要发送响应，直接返回: event=%#v\n", evt)
			} else {
				log.Printf("有函数调用需要发送响应，暂不返回\n")
			}
		case "response.output_text.delta":
			text := event.AsResponseOutputTextDelta()
			events <- domain.StreamEvent{
				ResponseID: responseID,
				Content:    text.Delta,
			}
		case "response.reasoning_summary_text.delta":
			r := event.AsResponseReasoningSummaryTextDelta()
			events <- domain.StreamEvent{
				ResponseID: responseID,
				// ReasoningContent: map[int64]string{
				// 	r.SummaryIndex: r.Delta,
				// },
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

			var output responses.ResponseInputItemUnionParam
			resp, err := fn.Call(fCtx, fcall.Request{Args: []byte(fc.Arguments)})
			if err != nil {
				h.logger.Error("执行函数调用失败",
					elog.FieldCustomKeyValue("函数名", fc.Name),
					elog.FieldCustomKeyValue("参数值", fc.Arguments),
					elog.FieldErr(err),
				)
				output = responses.ResponseInputItemParamOfFunctionCallOutput(fc.CallID, err.Error())
			} else {
				output = responses.ResponseInputItemUnionParam{
					OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
						CallID: fc.CallID,
						Output: resp.Output,
						Status: resp.Status,
					},
				}
			}
			if fn.Name() == fcall.NameAskUser {
				events <- domain.StreamEvent{
					ResponseID: responseID,
					CallID:     fc.CallID,
					Content:    fCtx.GetAttachment(fcall.NameAskUser),
				}
			} else {
				funcCallOutputs = append(funcCallOutputs, output)
			}
		}

	}
	err := stream.Err()
	if err != nil {
		h.logger.Error("forward过程中出错", elog.FieldErr(err))
		log.Printf("forward过程中出错: %+v\n", err)
	}

	if len(funcCallOutputs) > 0 {
		req.Messages = nil
		req.PreviousResponseID = responseID
		params := h.newParams(req)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		h.forward(req, h.client.Responses.NewStreaming(ctx, params), events)
		cancel()
	}
}

func (h *Handler) newInputItem(text string) responses.ResponseInputItemUnionParam {
	return responses.ResponseInputItemUnionParam{
		OfInputMessage: &responses.ResponseInputItemMessageParam{
			Content: []responses.ResponseInputContentUnionParam{{
				OfInputText: &responses.ResponseInputTextParam{
					Text: text,
				},
			}},
			Role: "user",
		},
	}
}
