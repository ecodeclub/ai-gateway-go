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
	"context"
	"fmt"

	ai "github.com/ecodeclub/ai-gateway-go/api/proto/gen/chat/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

var _ ai.ServiceServer = &ChatServer{}

type ChatServer struct {
	svc *service.ChatService
	ai.UnimplementedServiceServer
	logger    *elog.Component
	streamHdl stream.Handler
}

func NewChatServer(svc *service.ChatService, hdl stream.Handler) *ChatServer {
	return &ChatServer{svc: svc,
		streamHdl: hdl,
		logger:    elog.DefaultLogger.With(elog.FieldComponent("grpc.ChatServer"))}
}

func (c *ChatServer) Save(ctx context.Context, request *ai.SaveRequest) (*ai.SaveResponse, error) {
	chat := request.GetChat()
	sn, err := c.svc.Save(ctx, domain.ChatV1{
		Title: chat.Title,
		Uid:   chat.Uid,
		Sn:    chat.Sn,
	})
	if err != nil {
		return &ai.SaveResponse{}, err
	}
	return &ai.SaveResponse{Sn: sn}, nil
}

func (c *ChatServer) List(ctx context.Context, req *ai.ListRequest) (*ai.ListResponse, error) {
	chat, err := c.svc.List(ctx, req.Uid, req.Limit, req.Offset)
	if err != nil {
		return &ai.ListResponse{}, err
	}
	return &ai.ListResponse{Chats: slice.Map(chat, func(idx int, src domain.ChatV1) *ai.Chat {
		return c.toChatV1(src)
	})}, nil
}

func (c *ChatServer) toChatV1(chat domain.ChatV1) *ai.Chat {
	return &ai.Chat{
		Sn:    chat.Sn,
		Title: chat.Title,
		Uid:   chat.Uid,
		Msgs: slice.Map(chat.History(), func(idx int, src domain.HistoryRecord) *ai.Message {
			return &ai.Message{
				Role:    src.Role,
				Content: src.Content,
			}
		}),
		Ctime: chat.Ctime.UnixMilli(),
	}
}

func (c *ChatServer) Detail(ctx context.Context, request *ai.DetailRequest) (*ai.DetailResponse, error) {
	chat, err := c.svc.Detail(ctx, request.GetSn())
	if err != nil {
		return nil, err
	}
	return &ai.DetailResponse{Chat: c.toChatV1(chat)}, nil
}

func (c *ChatServer) StreamV1(request *ai.StreamV1Request, resp ai.Service_StreamV1Server) error {
	chat, err := c.svc.Detail(resp.Context(), request.ChatSn)
	if err != nil {
		return fmt.Errorf("查找 Chat 详情失败 %w", err)
	}

	turn := &domain.Turn{
		Vars: map[string]any{
			"Input": request.Input.Content,
		},
		UserRun: &domain.UserRun{
			Content: request.Input.Content,
			Files:   request.Input.Files,
		},
		AssistantRun: &domain.AssistantRun{
			// 构建当前的步骤，一般来说步骤不会超过 4 个
			Steps: make([]*domain.Step, 0, 4),
		},
	}
	// 默认第一个步骤就是调用 LLM，这个假设不要轻易破坏了
	// 后续如果有别的可能，那么就要考虑扩展 invocation_config 与这个地方了
	turn.AssistantRun.StartLLMStep(request.InvocationConfigId)
	// 初始化当前这一轮对话
	chat.Turns = append(chat.Turns, turn)
	ctx := &domain.StreamContext{
		Ctx: resp.Context(),
		Sender: &StreamEventGRPCSender{
			server: resp,
			logger: c.logger.With(elog.FieldComponentName("grpc.StreamEventGRPCSender")),
		},
		Chat: chat,
	}
	return c.streamHdl.Stream(ctx)
}

func (c *ChatServer) Stream(request *ai.StreamRequest, resp ai.Service_StreamServer) error {
	panic("implement me")
}
