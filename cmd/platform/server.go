package main

import (
	ds "github.com/cohesion-org/deepseek-go"
	ai "github.com/ecodeclub/ai-gateway-go/api/gen/ai/v1"
	igrpc "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/platform/deepseek"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"github.com/gotomicro/ego/server/egrpc"
)

// Server 配置并返回一个 gRPC 服务器实例。
// 返回值:
// - server.Server: 配置好的 gRPC 服务器实例。
// 功能说明:
// - 从配置中获取 DeepSeek 的 token。
// - 创建一个新的 DeepSeek 客户端和处理程序。
// - 使用处理程序创建 AI 服务实例。
// - 加载并构建 gRPC 服务器配置。
// - 注册 AI 服务到 gRPC 服务器。
func Server() server.Server {
	token := econf.GetString("deepseek.token")
	// 创建 DeepSeek 客户端和处理程序
	handler := deepseek.NewHandler(ds.NewClient(token))
	// 创建 AI 服务实例
	svc := service.NewAIService(handler)
	// 加载并构建 gRPC 服务器配置
	build := egrpc.Load("grpc.server").Build()
	// 注册 AI 服务到 gRPC 服务器
	ai.RegisterAIServiceServer(build.Server, igrpc.NewServer(svc))
	return build
}

// main 启动应用程序并运行 gRPC 服务器。
// 功能说明:
// - 创建一个新的 Ego 应用实例。
// - 使用 Server 函数配置应用的服务。
// - 启动应用并监听错误，如果启动失败则记录 panic 日志。
func main() {
	if err := ego.New().Serve(Server()).Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}
