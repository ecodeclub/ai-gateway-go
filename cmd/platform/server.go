package main

import (
	ds "github.com/cohesion-org/deepseek-go"
	pb "github.com/ecodeclub/ai-gateway-go/api/proto/gen/api/proto"
	igrpc "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/platform/deepseek"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"github.com/gotomicro/ego/server/egrpc"
	"google.golang.org/grpc"
	"net"
	"os"
)

func Server() server.Server {
	token := econf.GetString("deepseek.token")
	handler := deepseek.NewHandler(ds.NewClient(token))
	svc := service.NewAIService(handler)
	build := egrpc.Load("grpc.server").Build()
	pb.RegisterAIServiceServer(build.Server, igrpc.NewServer(svc))
	return build
}

func NewGrpcServer(port string) error {
	token := os.Getenv("token")
	handler := deepseek.NewHandler(ds.NewClient(token))
	svc := service.NewAIService(handler)
	grpcServer := igrpc.NewServer(svc)
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
	if err := ego.New().Serve(Server()).Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}
