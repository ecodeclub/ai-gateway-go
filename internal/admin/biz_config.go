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

	bg.POST("/save", ginx.BS[BizConfig](h.SaveBizConfig))
	bg.POST("/detail", ginx.BS[GetBizConfigReq](h.GetBizConfig))
	bg.POST("/list", ginx.BS[ListReq](h.List))
}

func (h *BizConfigHandler) SaveBizConfig(ctx *ginx.Context, req BizConfig, sess session.Session) (ginx.Result, error) {
	// 还不支持组织
	config := domain.BizConfig{
		ID:        req.ID,
		Name:      req.Name,
		OwnerID:   sess.Claims().Uid,
		OwnerType: domain.OwnerTypeUser.String(),
		Config:    req.Config,
	}

	id, err := h.svc.Save(ctx, config)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Code: 0,
		Msg:  "OK",
		Data: id,
	}, nil
}

type GetBizConfigReq struct {
	ID int64 `json:"id"`
}

func (h *BizConfigHandler) GetBizConfig(ctx *ginx.Context, req GetBizConfigReq, _ session.Session) (ginx.Result, error) {
	config, err := h.svc.GetByID(ctx.Request.Context(), req.ID)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg:  "OK",
		Data: newBizConfig(config),
	}, nil
}

type UpdateBizConfigReq struct {
	ID        int64  `json:"id"`
	OwnerId   int64  `json:"owner_id"`
	OwnerType string `json:"owner_type"`
	Config    string `json:"config"`
}

type DeleteBizConfigReq struct {
	ID int64 `json:"id"`
}

func (h *BizConfigHandler) List(ctx *ginx.Context, req ListReq, sess session.Session) (ginx.Result, error) {
	res, total, err := h.svc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: ginx.DataList[BizConfig]{
			List: slice.Map(res, func(idx int, src domain.BizConfig) BizConfig {
				return newBizConfig(src)
			}),
			Total: total,
		},
	}, err
}
