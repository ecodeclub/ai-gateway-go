//go:build wireinject

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
	"github.com/ecodeclub/ai-gateway-go/cmd/platform/ioc"
	"github.com/google/wire"
)

// InitApp 总体上这里的 app 都是从 platform 下的 ioc 复制过来的
// 只有一些需要 mock 的组件，才会作为参数传递进去
func InitApp(to TestOnly) *ioc.App {
	wire.Build(
		InitGin,
		InitDB,
		InitRedis,
		ioc.InitGrpcServer,
		ioc.MockSet,
		ioc.ChatSet,
		ioc.InvocationConfigSet,
		ioc.BizConfigSet,
		ioc.ProviderSet,
		ioc.ModelSet,
		wire.FieldsOf(new(TestOnly), "LLM"),
		wire.Struct(new(ioc.App), "*"),
	)
	return new(ioc.App)
}
