// Copyright 2025 ecodeclub
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

package ioc

import (
	"github.com/ecodeclub/ai-gateway-go/internal/admin"
	igrpc "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/cache"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/google/wire"
	"github.com/gotomicro/ego/server/egin"
	"github.com/gotomicro/ego/server/egrpc"
)

var (
	BaseSet = wire.NewSet(
		InitRedis, InitDB, InitSession,
		InitGin,
		InitGrpcServer)
	LLMSet  = wire.NewSet(initLLMHandler)
	ChatSet = wire.NewSet(
		dao.NewChatDAO,
		cache.NewChatCache,
		repository.NewChatRepo,
		service.NewChatService,
		igrpc.NewChatServer,
	)
	InvocationConfigSet = wire.NewSet(
		dao.NewInvocationConfigDAO,
		repository.NewInvocationConfigRepo,
		service.NewInvocationConfigService,
		admin.NewInvocationConfigHandler,
	)
	BizConfigSet = wire.NewSet(
		dao.NewBizConfigDAO,
		repository.NewBizConfigRepository,
		service.NewBizConfigService,
		admin.NewBizConfigHandler,
	)

	ProviderSet = wire.NewSet(
		dao.NewProviderDAO,
		repository.NewProviderRepository,
		InitProvider,
		admin.NewProviderHandler,
	)
	MockSet = wire.NewSet(admin.NewMockHandler)
)

type App struct {
	GrpcSever *egrpc.Component
	GinServer *egin.Component
}
