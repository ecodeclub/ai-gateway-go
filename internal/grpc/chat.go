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
	"errors"

	"github.com/ecodeclub/ai-gateway-go/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ai "github.com/ecodeclub/ai-gateway-go/api/proto/gen/chat/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ekit/slice"
)

var _ ai.ServiceServer = &ChatServer{}

type ChatServer struct {
	svc *service.ChatService
	ai.UnimplementedServiceServer
}

func NewChatServer(svc *service.ChatService) *ChatServer {
	return &ChatServer{svc: svc}
}

func (c *ChatServer) Save(ctx context.Context, request *ai.SaveRequest) (*ai.SaveResponse, error) {
	chat := request.GetChat()
	sn, err := c.svc.Save(ctx, domain.Chat{
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
	return &ai.ListResponse{Chats: slice.Map(chat, func(idx int, src domain.Chat) *ai.Chat {
		return c.toChat(src)
	})}, nil
}

func (c *ChatServer) toChat(chat domain.Chat) *ai.Chat {
	return &ai.Chat{
		Sn:    chat.Sn,
		Title: chat.Title,
		Uid:   chat.Uid,
		Msgs: slice.Map(chat.Messages, func(idx int, src domain.Message) *ai.Message {
			return &ai.Message{
				Role:             src.Role,
				Content:          src.Content,
				ReasoningContent: src.ReasoningContent,
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
	return &ai.DetailResponse{Chat: c.toChat(chat)}, nil
}

func (c *ChatServer) Stream(request *ai.StreamRequest, resp ai.Service_StreamServer) error {
	ctx := resp.Context()
	req := domain.ChatStreamRequest{
		Sn: request.GetSn(),
		Messages: slice.Map([]*ai.Message{request.GetMsg()}, func(idx int, src *ai.Message) domain.Message {
			return domain.Message{
				Role:    src.Role,
				Content: src.Content,
			}
		}),
		InvocationConfigID: request.GetInvocationConfigId(),
		Uid:                request.GetUid(),
		Key:                request.GetKey(),
		PreviousResponseID: request.GetPreviousResponseId(),
	}
	events, err := c.svc.Stream(ctx, req)
	if err != nil {
		if errors.Is(err, errs.ErrAccountOverdue) {
			return status.Error(codes.PermissionDenied, "账户欠费")
		}
		return err
	}
	return c.stream(ctx, events, resp)
}

func (c *ChatServer) stream(ctx context.Context, events chan domain.StreamEvent, resp ai.Service_StreamServer) error {
	var err error
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt, ok := <-events:
			if !ok || evt.Done {
				err = resp.Send(&ai.StreamResponse{Final: true})
				return err
			}
			if evt.Error != nil {
				err = resp.Send(&ai.StreamResponse{Err: evt.Error.Error()})
				return err
			}
			err = resp.Send(&ai.StreamResponse{
				ReasoningContent: evt.ReasoningContent,
				Content:          evt.Content,
			})
			if err != nil {
				return err
			}
		}
	}
}
