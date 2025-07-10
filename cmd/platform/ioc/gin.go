// Copyright 2023 ecodeclub
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
	"github.com/ecodeclub/ginx/session"
	"github.com/gotomicro/ego/server/egin"
)

func InitGin(
	sp session.Provider,
	promptHandler *admin.PromptHandler,
	bizConfig *admin.BizConfigHandler,
) *egin.Component {
	session.SetDefaultProvider(sp)
	res := egin.Load("admin").Build()
	// 登录校验
	res.Use(session.CheckLoginMiddleware())
	promptHandler.PrivateRoutes(res)
	bizConfig.PrivateRoutes(res)
	return res
}
