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
	chatv1 "github.com/ecodeclub/ai-gateway-go/api/proto/gen/chat/v1"
	igrpc "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/cache"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/openai"
	"github.com/ego-component/egorm"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"github.com/gotomicro/ego/server/egrpc"
	"github.com/redis/go-redis/v9"
)

func Server() server.Server {

	db := initDB()
	rdb := initRedis()
	handler := initLLMHandler()
	// 后期考虑用 wire，不过不得不说的是，wire 也是一坨屎
	grpcComponent := egrpc.Load("grpc.server").Build()
	chatv1.RegisterServiceServer(grpcComponent.Server,
		igrpc.NewServer(initChatService(db, rdb, handler)))
	return grpcComponent
}

func initLLMHandler() llm.Handler {
	type AliyunConfig struct {
		APIKey  string `json:"apiKey"`
		BaseURL string `json:"baseURL"`
		Model   string `json:"model"`
	}
	var cfg AliyunConfig

	err := econf.UnmarshalKey("aliyun", &cfg)
	if err != nil {
		panic(err)
	}
	handler := openai.NewHandler(cfg.APIKey, cfg.BaseURL, cfg.Model)
	return handler
}

func initChatService(db *egorm.Component,
	rdb redis.Cmdable,
	handler llm.Handler) *service.ChatService {
	chatDAO := dao.NewChatDAO(db)
	chatCache := cache.NewChatCache(rdb)
	repo := repository.NewChatRepo(chatDAO, chatCache)
	svc := service.NewChatService(repo, handler)
	return svc
}

// --config=local.yaml，替换你的配置文件地址
func main() {
	if err := ego.New().Serve(Server()).Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}
