package main

import (
	ds "github.com/cohesion-org/deepseek-go"
	pb "github.com/ecodeclub/ai-gateway-go/api/ai"
	GRPC "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/platform/deepseek"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"github.com/gotomicro/ego/server/egrpc"
	"os"
)

var (
	port  = "8080"
	token = os.Getenv("DEEPSEEK_TOKEN")
)

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
