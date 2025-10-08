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

package grpc

import (
	ai "github.com/ecodeclub/ai-gateway-go/api/proto/gen/chat/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/gotomicro/ego/core/elog"
)

type StreamEventGRPCSender struct {
	server ai.Service_StreamV1Server
	logger *elog.Component
}

func (sender *StreamEventGRPCSender) Send(evt domain.StreamEventV1) error {
	switch {
	case evt.Delta != nil:
		return sender.server.Send(&ai.StreamV1Response{
			Event: &ai.StreamV1Response_Delta{Delta: &ai.Delta{Content: evt.Delta.Content}},
		})
	case evt.Usage != nil:
		return sender.server.Send(&ai.StreamV1Response{
			Event: &ai.StreamV1Response_Usage{
				Usage: &ai.Usage{
					InputTokens:  evt.Usage.InputTokens,
					OutputTokens: evt.Usage.OutputTokens,
					TotalTokens:  evt.Usage.TotalTokens,
				},
			},
		})

	case evt.StepUpdate != nil:
		return sender.server.Send(&ai.StreamV1Response{
			Event: &ai.StreamV1Response_StepUpdate{StepUpdate: &ai.StepUpdate{
				Name:    evt.StepUpdate.Title,
				Summary: evt.StepUpdate.Summary,
			}},
		})
	case evt.Err != nil:
		return sender.server.Send(&ai.StreamV1Response{
			Event: &ai.StreamV1Response_Error{
				Error: &ai.Error{
					Message: evt.Err.Error(),
				},
			},
		})
	default:
		msg := "未知的事件发生"
		sender.logger.Error(msg, elog.Any("evt", evt))
		return sender.server.Send(&ai.StreamV1Response{
			Event: &ai.StreamV1Response_Error{
				Error: &ai.Error{
					Message: msg,
				},
			},
		})
	}
}
