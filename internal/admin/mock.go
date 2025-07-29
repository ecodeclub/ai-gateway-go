//go:build !product

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

package admin

import (
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gotomicro/ego/server/egin"
)

// MockHandler 这是 mock 的东西，你在生产环境要关掉
type MockHandler struct {
}

func NewMockHandler() *MockHandler {
	return new(MockHandler)
}

func (h *MockHandler) PublicRoutes(server *egin.Component) {
	server.Any("/mock/login", ginx.W(h.MockLogin))
}

func (h *MockHandler) MockLogin(ctx *ginx.Context) (ginx.Result, error) {
	const uid = 1
	// 构建session
	jwtData := map[string]string{}
	_, err := session.NewSessionBuilder(ctx, uid).SetJwtData(jwtData).Build()
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
		Data: Profile{
			Nickname: "模拟用户",
		},
	}, nil
}

type Profile struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}
