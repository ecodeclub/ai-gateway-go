package main

import (
	ds "github.com/cohesion-org/deepseek-go"
	ai "github.com/ecodeclub/ai-gateway-go/api/gen/ai/v1"
	igrpc "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/grpc/middleware"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/platform/deepseek"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server/egrpc"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	if err := ego.New().Serve(Server()).Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}

func Server() *egrpc.Component {
	// 初始化数据库
	db := initDB()

	// 初始化 BizConfig 服务
	bizConfigService := InitBizConfigService(db)

	// 生成 Auth 中间件
	authInterceptor := middleware.AuthInterceptor(bizConfigService)

	// 加载配置并追加拦截器（必须用 WithServerOption 包装）
	grpcContainer := egrpc.Load("grpc.server")
	build := grpcContainer.Build(egrpc.WithServerOption(grpc.UnaryInterceptor(authInterceptor)))

	// 注册 AI 服务
	token := econf.GetString("deepseek.token")
	handler := deepseek.NewHandler(ds.NewClient(token))
	aiSvc := service.NewAIService(handler)
	ai.RegisterAIServiceServer(build.Server, igrpc.NewServer(aiSvc))

	// 注册 BizConfig 服务
	bizConfigServer := igrpc.NewBizConfigServer(bizConfigService)
	ai.RegisterBizConfigServiceServer(build.Server, bizConfigServer)

	return build
}

// initDB 初始化数据库并自动建表
func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/bizconfig"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

// InitBizConfigService 初始化 BizConfigService 实例
func InitBizConfigService(db *gorm.DB) *service.BizConfigService {
	bizconfigdao := dao.NewBizConfigDAO(db)
	repo := repository.NewBizConfigRepository(bizconfigdao)
	svc := service.NewBizConfigService(
		repo,
		econf.GetString("jwt.secret"),
		econf.GetDuration("jwt.expire"),
	)
	return svc
}
