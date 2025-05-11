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

func main() {
	infra.Init()
	db := initDB()
	server := gin.Default()
	bizconfig := initBizConfig(db)
	bizconfig.RegisterRoutes(server)
	server.Run(":8080")
}

// initDB 初始化数据库并自动建表
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
func initBizConfig(db *gorm.DB) *web.BizConfigHandler {
	dao := dao.NewBizConfigDAO(db)
	repo := repository.NewBizConfigRepository(dao)
	svc := service.NewBizConfigService(repo)
	server := web.NewBizConfigHandler(svc)
	return server
}
