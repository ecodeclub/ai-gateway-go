package main

import (
	ds "github.com/cohesion-org/deepseek-go"
	pb "github.com/ecodeclub/ai-gateway-go/api/gen/api/proto"
	GRPC "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/platform/deepseek"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/server"
	"github.com/gotomicro/ego/server/egrpc"
	"google.golang.org/grpc"
	"net"
	"os"
)

var (
	token = os.Getenv("DEEPSEEK_TOKEN")
)

func DeepSeekServer() server.Server {
	handler := deepseek.NewHandler(ds.NewClient(token))
	svc := service.NewAIService(handler)
	build := egrpc.Load("").Build()
	pb.RegisterAIServiceServer(build.Server, GRPC.NewServer(svc))
	return build
}
func EgoStart() error {
	err := ego.New().Run()
	return err
}

func NewGrpcServer(port string) error {
	handler := deepseek.NewHandler(ds.NewClient(token))
	svc := service.NewAIService(handler)
	grpcServer := GRPC.NewServer(svc)
	listener, err := net.Listen("tcp", ":"+port)

	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterAIServiceServer(s, grpcServer)
	if err := s.Serve(listener); err != nil {
		return err
	}
	return nil
}

func main() {
	err := NewGrpcServer("8080")
	if err != nil {
	}
}
