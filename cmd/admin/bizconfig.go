package main

import (
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/web"
	"github.com/ecodeclub/ai-gateway-go/internal/web/infra"
	//"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// main 函数是程序的入口点
// 主要职责:
//  1. 初始化基础设施
//  2. 建立数据库连接
//  3. 创建并配置 Gin 框架实例
//  4. 注册路由
//  5. 启动 HTTP 服务器
func main() {
	infra.Init()
	db := initDB()
	server := gin.Default()
	bizconfig := initBizConfig(db)
	bizconfig.RegisterRoutes(server)
	server.Run(":8080")
}

// initDB 初始化数据库并自动建表
// 返回值:
//
//	*gorm.DB - 初始化后的数据库连接实例
func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/ai_gateway_platform"))
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
// 参数:
//
//	db *gorm.DB - 数据库连接实例
//
// 返回值:
//
//	*web.BizConfigHandler - 业务配置处理器实例
func initBizConfig(db *gorm.DB) *web.BizConfigHandler {
	dao := dao.NewBizConfigDAO(db)
	repo := repository.NewBizConfigRepository(dao)
	svc := service.NewBizConfigService(repo)
	server := web.NewBizConfigHandler(svc)
	return server
}
