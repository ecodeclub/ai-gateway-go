package main

import (
	"fmt"
	ds "github.com/cohesion-org/deepseek-go"
	pb "github.com/ecodeclub/ai-gateway-go/api/proto"
	GRPC "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/platform/deepseek"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"github.com/gotomicro/ego/server/egrpc"
	"google.golang.org/grpc"
	"net"
	"os"
)

var (
	port  = "8080"
	token = os.Getenv("DEEPSEEK_TOKEN")
)

func main_() {
	handler := deepseek.NewHandler(ds.NewClient(token))
	svc := service.NewAIService(handler)

	se := GRPC.NewServer(svc)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println(err)
	}

	s := grpc.NewServer()
	pb.RegisterAIServiceServer(s, se)
	if err := s.Serve(listener); err != nil {
		fmt.Println("", err)
	}
}

func AIServer() server.Server {
	handler := deepseek.NewHandler(ds.NewClient(token))
	svc := service.NewAIService(handler)
	build := egrpc.Load("server.grpc").Build()
	pb.RegisterAIServiceServer(build.Server, GRPC.NewServer(svc))
	return build
}

func main() {
	if err := ego.New().Serve(AIServer()).Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}
