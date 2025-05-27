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
