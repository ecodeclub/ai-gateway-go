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

package main

import (
	"github.com/ecodeclub/ai-gateway-go/cmd/platform/ioc"
	"github.com/gotomicro/ego"
)

// --config=local.yaml，替换你的配置文件地址
func main() {
	egoApp := ego.New()
	app := ioc.InitApp()

	err := egoApp.
		// Invoker 在 Ego 里面，应该叫做初始化函数
		Invoker().
		Serve(
			app.GinServer,
			app.GrpcSever).
		Run()
	panic(err)
}
