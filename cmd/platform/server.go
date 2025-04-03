package main

import (
	ds "github.com/cohesion-org/deepseek-go"
	ai "github.com/ecodeclub/ai-gateway-go/api/proto/gen"
	igrpc "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/platform/deepseek"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"github.com/gotomicro/ego/server/egrpc"
)

func Server() server.Server {
	token := econf.GetString("deepseek.token")
	handler := deepseek.NewHandler(ds.NewClient(token))
	svc := service.NewAIService(handler)
	build := egrpc.Load("grpc.server").Build()
	ai.RegisterAIServiceServer(build.Server, igrpc.NewServer(svc))
	return build
}

func main() {
	if err := ego.New().Serve(Server()).Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}
