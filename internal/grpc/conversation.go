package grpc

import (
	"context"

	ai "github.com/ecodeclub/ai-gateway-go/api/gen/ai/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ekit/slice"
)

type ConversationServer struct {
	svc *service.ConversationService
	ai.UnimplementedConversationServiceServer
}

func NewConversationServer(svc *service.ConversationService) *ConversationServer {
	return &ConversationServer{svc: svc}
}

func (c *ConversationServer) Create(ctx context.Context, conversation *ai.Conversation) (*ai.Conversation, error) {
	id, err := c.svc.Create(ctx, domain.Conversation{
		Sn:    conversation.Sn,
		Title: conversation.Title,
		Uid:   conversation.Uid,
	})

	if err != nil {
		return &ai.Conversation{}, err
	}
	return &ai.Conversation{Sn: id}, nil
}

func (c *ConversationServer) List(ctx context.Context, req *ai.ListReq) (*ai.ListResp, error) {
	conversation, err := c.svc.List(ctx, req.Uid, req.Limit, req.Offset)
	if err != nil {
		return &ai.ListResp{}, err
	}
	return &ai.ListResp{Conversations: c.toConversation(conversation)}, nil
}

func (c *ConversationServer) Chat(ctx context.Context, request *ai.LLMRequest) (*ai.ChatResponse, error) {
	response, err := c.svc.Chat(ctx, request.Sn, c.toDomainMessage(request.Message))

	if err != nil {
		return &ai.ChatResponse{}, err
	}

	return &ai.ChatResponse{
		Sn: request.Sn,
		Response: &ai.Message{
			Role:             ai.Role(response.Response.Role),
			Content:          response.Response.Content,
			ReasoningContent: response.Response.ReasoningContent,
		},
	}, nil
}

func (c *ConversationServer) Stream(request *ai.LLMRequest, resp ai.ConversationService_StreamServer) error {
	ctx := resp.Context()
	ch, err := c.svc.Stream(ctx, request.Sn, c.toDomainMessage(request.Message))
	if err != nil {
		return err
	}
	return c.stream(ctx, ch, resp)
}

func (c *ConversationServer) stream(ctx context.Context, ch chan domain.StreamEvent, resp ai.ConversationService_StreamServer) error {
	var err error
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e, ok := <-ch:
			if !ok || e.Done {
				err = resp.Send(&ai.StreamEvent{Final: true})
				return err
			}
			if e.Error != nil {
				err = resp.Send(&ai.StreamEvent{Err: e.Error.Error()})
				return err
			}
			err = resp.Send(&ai.StreamEvent{Final: false, Content: e.Content, ReasoningContent: e.ReasoningContent})
			if err != nil {
				return err
			}
		}
	}
}

func (c *ConversationServer) toMessage(messages []domain.Message) []*ai.Message {
	return slice.Map(messages, func(idx int, src domain.Message) *ai.Message {
		return &ai.Message{
			Role:             ai.Role(src.Role),
			Content:          src.Content,
			ReasoningContent: src.Content,
		}
	})
}

func (c *ConversationServer) toConversation(conversations []domain.Conversation) []*ai.Conversation {
	return slice.Map(conversations, func(idx int, src domain.Conversation) *ai.Conversation {
		return &ai.Conversation{
			Sn:    src.Sn,
			Title: src.Title,
		}
	})
}

func (c *ConversationServer) toDomainMessage(messages []*ai.Message) []domain.Message {
	return slice.Map(messages, func(idx int, src *ai.Message) domain.Message {
		return domain.Message{
			Role:    int64(src.Role),
			Content: src.Content,
		}
	})
}
