package grpc

import (
	"fmt"
	"github.com/ecodeclub/ai-gateway-go/api/gen"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"time"
)

type Server struct {
	svc *service.AIService
	gen.UnimplementedAIServiceServer
}

func NewServer(svc *service.AIService) *Server {
	return &Server{svc: svc}
}

func (server *Server) Ask(r *gen.AskRequest, s gen.AIService_AskServer) error {
	msg, err := server.svc.Ask(r.GetId(), r.GetText())
	if err != nil {
		return err
	}

	go func() {
		err := msg.Work()
		if err != nil {
			fmt.Println("msg 错误")
		}
	}()

	time.Sleep(time.Second)

	for content := range msg.Chan {
		if content == "finished" {
			return nil
		}

		res := &gen.AskResponse{
			Text: content,
		}

		if err := s.Send(res); err != nil {
			return err
		}
	}
	return nil
}
