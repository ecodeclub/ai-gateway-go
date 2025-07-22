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
	g.POST("/basic", ginx.BS(h.Basic))
	g.POST("/versions/list", ginx.BS(h.ListVersions))
	g.POST("/versions/save", ginx.BS(h.SaveVersion))
	g.POST("/versions/detail", ginx.BS[IDReq](h.VersionDetail))
	g.POST("/versions/activate", ginx.BS[IDReq](h.ActivateVersion))
	g.POST("/fork", ginx.B(h.Fork))
}

func (h *InvocationConfigHandler) PublicRoutes(_ *gin.Engine) {}

func (h *InvocationConfigHandler) Save(ctx *ginx.Context, req SaveInvocationConfigReq, sess session.Session) (ginx.Result, error) {
	id, err := h.svc.Save(ctx, req.Cfg.toDomain())
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg:  "OK",
		Data: id,
	}, nil
}

func (h *InvocationConfigHandler) Basic(ctx *ginx.Context, req IDReq, sess session.Session) (ginx.Result, error) {
	res, err := h.svc.Get(ctx, req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: newInvocationVO(res),
	}, nil
}

// Delete 删除整个 prompt
func (h *InvocationConfigHandler) Delete(ctx *ginx.Context, req DeleteReq) (ginx.Result, error) {
	err := h.svc.Delete(ctx, req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *InvocationConfigHandler) DeleteVersion(ctx *ginx.Context, req DeleteVersionReq) (ginx.Result, error) {
	err := h.svc.DeleteVersion(ctx, req.VersionID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// UpdatePrompt 更新 prompt 的基本信息
func (h *InvocationConfigHandler) UpdatePrompt(ctx *ginx.Context, req UpdatePromptReq) (ginx.Result, error) {
	prompt := domain.InvocationConfig{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
	}
	err := h.svc.UpdateInfo(ctx, prompt)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *InvocationConfigHandler) UpdateVersion(ctx *ginx.Context, req UpdateVersionReq) (ginx.Result, error) {
	version := domain.InvocationCfgVersion{
		ID:           req.VersionID,
		Prompt:       req.Content,
		SystemPrompt: req.SystemContent,
		Temperature:  req.Temperature,
		TopP:         req.TopN,
		MaxTokens:    req.MaxTokens,
	}
	err := h.svc.UpdateVersion(ctx, version)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *InvocationConfigHandler) Publish(ctx *ginx.Context, req PublishReq) (ginx.Result, error) {
	err := h.svc.Publish(ctx, req.VersionID, req.Label)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// Fork 新增一个版本
func (h *InvocationConfigHandler) Fork(ctx *ginx.Context, req ForkReq) (ginx.Result, error) {
	id, err := h.svc.Fork(ctx, req.VersionID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg:  "OK",
		Data: id,
	}, nil
}

func (h *InvocationConfigHandler) List(ctx *ginx.Context, req ListInvocationConfigReq, sess session.Session) (ginx.Result, error) {
	cfgs, total, err := h.svc.List(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: ginx.DataList[InvocationConfigVO]{
			List: slice.Map(cfgs, func(idx int, src domain.InvocationConfig) InvocationConfigVO {
				return newInvocationVO(src)
			}),
			Total: total,
		},
	}, nil
}

func (h *InvocationConfigHandler) ListVersions(ctx *ginx.Context, req ListInvocationConfigVersionsReq, sess session.Session) (ginx.Result, error) {
	list, total, err := h.svc.ListVersions(ctx, req.InvID, req.Offset, req.Limit)
	return ginx.Result{
		Data: ginx.DataList[InvocationCfgVersionVO]{
			List: slice.Map(list, func(idx int, src domain.InvocationCfgVersion) InvocationCfgVersionVO {
				return newInvocationCfgVersion(src)
			}),
			Total: total,
		},
	}, err
}

func (h *InvocationConfigHandler) SaveVersion(ctx *ginx.Context, req SaveInvocationConfigVersionReq, sess session.Session) (ginx.Result, error) {
	id, err := h.svc.SaveVersion(ctx, req.Version.toDomain())
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg:  "OK",
		Data: id,
	}, nil
}

func (h *InvocationConfigHandler) VersionDetail(ctx *ginx.Context, req IDReq, sess session.Session) (ginx.Result, error) {
	res, err := h.svc.GetVersion(ctx, req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: newInvocationCfgVersion(res),
	}, nil
}

func (h *InvocationConfigHandler) ActivateVersion(ctx *ginx.Context, req IDReq, sess session.Session) (ginx.Result, error) {
	err := h.svc.ActivateVersion(ctx, req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}
