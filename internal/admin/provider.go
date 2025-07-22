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

package admin

import (
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gotomicro/ego/server/egin"
)

type ProviderHandler struct {
}

func NewProviderHandler() *ProviderHandler {
	return &ProviderHandler{}
}

func (h *ProviderHandler) PrivateRoutes(server *egin.Component) {
	g := server.Group("/providers")
	g.POST("/all", ginx.S(h.AllProviders))
}

func (h *ProviderHandler) AllProviders(ctx *ginx.Context, sess session.Session) (ginx.Result, error) {
	return ginx.Result{
		Data: ginx.DataList[ProviderVO]{
			List: []ProviderVO{
				{
					ID:   1,
					Name: "阿里百炼",
					Models: []ModelVO{
						{
							ID:   1,
							Name: "deepseek-r1-a",
						},
						{
							ID:   2,
							Name: "deepseek-r1-a",
						},
					},
				},
				{
					ID:   2,
					Name: "百度千帆",
					Models: []ModelVO{
						{
							ID:   3,
							Name: "deepseek-r1-b",
						},
						{
							ID:   4,
							Name: "deepseek-r1-b",
						},
					},
				},
			},
		},
	}, nil
}
