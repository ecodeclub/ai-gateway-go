package grpc

import (
	"context"
	ai "github.com/ecodeclub/ai-gateway-go/api/gen/ai/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
)

// Server 是一个 gRPC 服务器结构体，实现了 ai.AIServiceServer 接口。
// 它包含一个内部的服务实例和一个未实现的 gRPC 服务接口。
type Server struct {
	svc *service.AIService
	ai.UnimplementedAIServiceServer
}

// NewServer 创建一个新的 Server 实例。
// 参数:
//
//	svc: 内部服务实例。
//
// 返回值:
//
//	指向新创建的 Server 实例的指针。
func NewServer(svc *service.AIService) *Server {
	return &Server{svc: svc}
}

// Invoke 处理来自客户端的同步调用请求。
// 参数:
//
//	ctx: 上下文对象，用于控制请求生命周期。
//	r:   LLMRequest 请求对象，包含请求的数据。
//
// 返回值:
//
//	LLMResponse 响应对象，包含处理结果或错误信息。
func (server *Server) Invoke(ctx context.Context, r *ai.LLMRequest) (*ai.LLMResponse, error) {
	resp, err := server.svc.Invoke(
		ctx,
		domain.LLMRequest{Id: r.GetId(), Text: r.GetText()})

	if err != nil {
		return &ai.LLMResponse{}, err
	}

	return &ai.LLMResponse{Content: resp.Content}, nil
}

// Stream 处理流式请求。
// 参数:
//
//	r:   LLMRequest 请求对象，包含请求的数据。
//	resp: AIService_StreamServer 接口，用于发送流式响应。
//
// 返回值:
//
//	错误信息（如果有）。
func (server *Server) Stream(r *ai.LLMRequest, resp ai.AIService_StreamServer) error {
	ctx := resp.Context()

	ch, err := server.svc.Stream(
		ctx,
		domain.LLMRequest{Id: r.GetId(), Text: r.GetText()})

	if err != nil {
		return err
	}

	return server.stream(ctx, ch, resp)
}

// stream 发送流式响应给客户端。
// 参数:
//
//	ctx: 上下文对象，用于控制请求生命周期。
//	ch:  流事件通道，用于接收流事件。
//	resp: AIService_StreamServer 接口，用于发送流式响应。
//
// 返回值:
//
//	错误信息（如果有）。
func (server *Server) stream(ctx context.Context, ch chan domain.StreamEvent, resp ai.AIService_StreamServer) error {
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
			err = resp.Send(&ai.StreamEvent{Final: false, Content: e.Content})
			if err != nil {
				return err
			}
		}
	}
}
