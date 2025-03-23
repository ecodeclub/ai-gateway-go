package grpc

import (
	"fmt"
	pb "github.com/ecodeclub/ai-gateway-go/pkg/proto"
	"github.com/ecodeclub/ai-gateway-go/service"
	"os"
	"time"
)

var token = os.Getenv("DEEPSEEK_API_KEY")

type Server struct {
	pb.UnimplementedAIServiceServer
}

func (server *Server) Ask(r *pb.AskRequest, s pb.AIService_AskServer) error {
	msg, err := service.SendReq(token, r.GetText())
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

		// 未来考虑加上更多的返回字段
		res := &pb.AskResponse{
			Text: content,
		}

		if err := s.Send(res); err != nil {
			return err
		}
	}
	return nil
}
