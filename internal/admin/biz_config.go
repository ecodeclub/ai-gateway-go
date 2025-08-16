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
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gotomicro/ego/server/egin"
)

type BizConfigHandler struct {
	svc *service.BizConfigService
}

func NewBizConfigHandler(svc *service.BizConfigService) *BizConfigHandler {
	return &BizConfigHandler{svc: svc}
}

func (h *BizConfigHandler) PrivateRoutes(server *egin.Component) {
	bg := server.Group("/biz-configs")
	bg.POST("/save", ginx.BS(h.Save))
	bg.POST("/list", ginx.BS(h.List))
	bg.POST("/detail", ginx.BS(h.Detail))
}

func (h *BizConfigHandler) Save(ctx *ginx.Context, req BizConfig, _ session.Session) (ginx.Result, error) {
	id, err := h.svc.Save(ctx, domain.BizConfig{
		ID:        req.ID,
		Name:      req.Name,
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Config:    req.Config,
	})
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg:  "OK",
		Data: id,
	}, nil
}

func (h *BizConfigHandler) List(ctx *ginx.Context, req ListReq, _ session.Session) (ginx.Result, error) {
	res, total, err := h.svc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: ginx.DataList[BizConfig]{
			List: slice.Map(res, func(_ int, src domain.BizConfig) BizConfig {
				return newBizConfig(src)
			}),
			Total: int(total),
		},
	}, err
}

func (h *BizConfigHandler) Detail(ctx *ginx.Context, req IDReq, _ session.Session) (ginx.Result, error) {
	config, err := h.svc.Detail(ctx.Request.Context(), req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg:  "OK",
		Data: newBizConfig(config),
	}, nil
}
