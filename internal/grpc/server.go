package grpc

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ai-gateway-go/api/ai"
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
	ch, err := server.svc.Stream(
		context.Background(),
		domain.StreamRequest{Id: r.GetId(), Text: r.GetText()})

	if err != nil {
		return err
	}

	err = server.stream(ch, resp)
	if err != nil {
		return err
	}

	return nil
}

func (server *Server) stream(ch chan domain.StreamEvent, resp ai.AIService_StreamServer) error {
	var err error
	for {
		select {
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
		}
	}
}
