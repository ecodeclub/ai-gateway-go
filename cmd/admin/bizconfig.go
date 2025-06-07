// Copyright 2021 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/web"
	"github.com/ecodeclub/ai-gateway-go/internal/web/infra"

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
	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
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
	dao1 := dao.NewBizConfigDAO(db)
	repo := repository.NewBizConfigRepository(dao1)
	svc := service.NewBizConfigService(repo)
	server := web.NewBizConfigHandler(svc)
	return server
}
