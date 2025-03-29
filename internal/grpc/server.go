package grpc

import (
	"context"
	"fmt"
	ai "github.com/ecodeclub/ai-gateway-go/api/gen/api/proto"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
)

type Server struct {
	svc *service.AIService
	ai.UnimplementedAIServiceServer
}

func NewServer(svc *service.AIService) *Server {
	return &Server{svc: svc}
}

func (server *Server) Stream(r *ai.StreamRequest, resp ai.AIService_StreamServer) error {
	ctx := context.Background()

	ch, err := server.svc.Stream(
		ctx,
		domain.StreamRequest{Id: r.GetId(), Text: r.GetText()})

	if err != nil {
		return err
	}

	err = server.stream(ctx, ch, resp)
	if err != nil {
		return err
	}

	return nil
}

func (server *Server) stream(ctx context.Context, ch chan domain.StreamEvent, resp ai.AIService_StreamServer) error {
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
				err = resp.Send(&ai.StreamResponse{Err: fmt.Sprint(e.Error)})
				return err
			}
			err = resp.Send(&ai.StreamResponse{Final: false, Text: e.Content})
			if err != nil {
				return err
			}
		}
	}
}
