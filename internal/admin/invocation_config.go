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
	"fmt"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/server/egin"
)

type InvocationConfigHandler struct {
	svc *service.InvocationConfigService
}

func NewInvocationConfigHandler(svc *service.InvocationConfigService) *InvocationConfigHandler {
	res := &InvocationConfigHandler{svc: svc}
	return res
}

func (h *InvocationConfigHandler) PrivateRoutes(server *egin.Component) {
	g := server.Group("/invocation-configs")
	g.POST("/save", ginx.BS(h.Save))
	g.POST("/list", ginx.BS(h.List))
	g.POST("/detail", ginx.BS(h.Detail))
	g.POST("/versions/save", ginx.BS(h.SaveVersion))
	g.POST("/versions/list", ginx.BS(h.ListVersions))
	g.POST("/versions/detail", ginx.BS[IDReq](h.VersionDetail))
	g.POST("/versions/activate", ginx.BS[IDReq](h.ActivateVersion))
	g.POST("/versions/fork", ginx.B(h.ForkVersion))
}

func (h *InvocationConfigHandler) PublicRoutes(_ *gin.Engine) {}

func (h *InvocationConfigHandler) Save(ctx *ginx.Context, req SaveInvocationConfigReq, _ session.Session) (ginx.Result, error) {
	id, err := h.svc.Save(ctx.Request.Context(), req.Cfg.toDomain())
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg:  "OK",
		Data: id,
	}, nil
}

func (h *InvocationConfigHandler) List(ctx *ginx.Context, req ListInvocationConfigReq, _ session.Session) (ginx.Result, error) {
	cfgs, total, err := h.svc.List(ctx.Request.Context(), req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: ginx.DataList[InvocationConfigVO]{
			List: slice.Map(cfgs, func(_ int, src domain.InvocationConfig) InvocationConfigVO {
				return newInvocationVO(src)
			}),
			Total: total,
		},
	}, nil
}

func (h *InvocationConfigHandler) Detail(ctx *ginx.Context, req IDReq, _ session.Session) (ginx.Result, error) {
	res, err := h.svc.Detail(ctx.Request.Context(), req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: newInvocationVO(res),
	}, nil
}

func (h *InvocationConfigHandler) SaveVersion(ctx *ginx.Context, req SaveInvocationConfigVersionReq, _ session.Session) (ginx.Result, error) {
	version := req.Version.toDomain()
	if !version.Status.IsValid() {
		return systemErrorResult, fmt.Errorf("版本状态非法：%q", version.Status.String())
	}
	id, err := h.svc.SaveVersion(ctx.Request.Context(), version)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg:  "OK",
		Data: id,
	}, nil
}

func (h *InvocationConfigHandler) ListVersions(ctx *ginx.Context, req ListInvocationConfigVersionsReq, _ session.Session) (ginx.Result, error) {
	list, total, err := h.svc.ListVersions(ctx.Request.Context(), req.InvID, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: ginx.DataList[InvocationConfigVersionVO]{
			List: slice.Map(list, func(idx int, src domain.InvocationConfigVersion) InvocationConfigVersionVO {
				return newInvocationCfgVersion(src)
			}),
			Total: total,
		},
	}, nil
}

func (h *InvocationConfigHandler) VersionDetail(ctx *ginx.Context, req IDReq, _ session.Session) (ginx.Result, error) {
	res, err := h.svc.VersionDetail(ctx.Request.Context(), req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: newInvocationCfgVersion(res),
	}, nil
}

func (h *InvocationConfigHandler) ActivateVersion(ctx *ginx.Context, req IDReq, _ session.Session) (ginx.Result, error) {
	err := h.svc.ActivateVersion(ctx.Request.Context(), req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *InvocationConfigHandler) ForkVersion(ctx *ginx.Context, req IDReq) (ginx.Result, error) {
	id, err := h.svc.ForkVersion(ctx.Request.Context(), req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg:  "OK",
		Data: id,
	}, nil
}
