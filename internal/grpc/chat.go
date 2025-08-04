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
	chatSvc := &ChatServer{svc: svc}
	return chatSvc
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
	return &ai.ListResponse{Chats: c.toChats(chat)}, nil
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
	ch, err := c.svc.Stream(
		ctx,
		request.GetSn(),
		request.GetUid(),
		request.GetModelId(),
		c.toDomainMessage([]*ai.Message{request.GetMsg()}))
	if err != nil {
		if errors.Is(err, errs.ErrAccountOverdue) {
			return status.Error(codes.PermissionDenied, "账户欠费")
		}
		return err
	}
	return c.stream(ctx, ch, resp)
}

func (c *ChatServer) stream(ctx context.Context, ch chan domain.StreamEvent, resp ai.Service_StreamServer) error {
	var err error
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e, ok := <-ch:
			if !ok || e.Done {
				err = resp.Send(&ai.StreamResponse{Final: true})
				return err
			}
			if e.Error != nil {
				err = resp.Send(&ai.StreamResponse{Err: e.Error.Error()})
				return err
			}
			err = resp.Send(&ai.StreamResponse{Final: false, Content: e.Content, ReasoningContent: e.ReasoningContent})
			if err != nil {
				return err
			}
		}
	}
}

func (c *ChatServer) toChats(conversations []domain.Chat) []*ai.Chat {
	return slice.Map(conversations, func(idx int, src domain.Chat) *ai.Chat {
		return c.toChat(src)
	})
}

func (c *ChatServer) toChat(chat domain.Chat) *ai.Chat {
	return &ai.Chat{
		Sn:    chat.Sn,
		Title: chat.Title,
		Uid:   chat.Uid,
		Msgs:  c.toMessage(chat.Messages),
		Ctime: chat.Ctime.UnixMilli(),
	}
}

func (c *ChatServer) toDomainMessage(messages []*ai.Message) []domain.Message {
	return slice.Map(messages, func(idx int, src *ai.Message) domain.Message {
		return domain.Message{
			Role:    src.Role,
			Content: src.Content,
		}
	})
}

func (c *ChatServer) toMessage(messages []domain.Message) []*ai.Message {
	return slice.Map(messages, func(idx int, src domain.Message) *ai.Message {
		return &ai.Message{
			Role:             src.Role,
			Content:          src.Content,
			ReasoningContent: src.ReasoningContent,
		}
	})
}
