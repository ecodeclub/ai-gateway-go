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

package web

import (
	"errors"
	"strconv"

	"github.com/ecodeclub/ai-gateway-go/errs"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

type BizConfigHandler struct {
	svc service.BizConfigService
}

func NewBizConfigHandler(svc service.BizConfigService) *BizConfigHandler {
	return &BizConfigHandler{svc: svc}
}

func (h *BizConfigHandler) RegisterRoutes(server *gin.Engine) {
	bg := server.Group("/api/v1/biz-configs")

	bg.POST("/create", ginx.BS[CreateBizConfigReq](h.CreateBizConfig))
	bg.POST("/get", ginx.BS[GetBizConfigReq](h.GetBizConfig))
	bg.POST("/update", ginx.BS[UpdateBizConfigReq](h.UpdateBizConfig))
	bg.POST("/delete", ginx.BS[DeleteBizConfigReq](h.DeleteBizConfig))
}

type CreateBizConfigReq struct {
	ID        int64  `json:"id"`
	OwnerId   int64  `json:"owner_id"`
	OwnerType string `json:"owner_type"`
	Config    string `json:"config"`
}

func (h *BizConfigHandler) CreateBizConfig(ctx *ginx.Context, req CreateBizConfigReq, _ session.Session) (ginx.Result, error) {
	config := domain.BizConfig{
		ID:        req.ID,
		OwnerID:   req.OwnerId,
		OwnerType: req.OwnerType,
		Config:    req.Config,
	}

	created, err := h.svc.Create(ctx.Request.Context(), config)
	if err != nil {
		return ginx.Result{Code: 500, Msg: "failed to create biz config"}, err
	}

	return ginx.Result{
		Code: 0,
		Msg:  "success",
		Data: gin.H{
			"bizconfig": h.toResponse(created),
		},
	}, nil
}

type BizConfig struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	OwnerId   int64  `json:"ownerID"`
	OwnerType string `json:"ownerType"`
	Config    string `json:"config"`
}

type GetBizConfigReq struct {
	ID int64 `json:"id"`
}

func (h *BizConfigHandler) GetBizConfig(ctx *ginx.Context, req GetBizConfigReq, _ session.Session) (ginx.Result, error) {
	config, err := h.svc.GetByID(ctx.Request.Context(), req.ID)
	if err == errs.ErrBizConfigNotFound {
		return ginx.Result{Code: 404, Msg: "biz config not found"}, nil
	} else if err != nil {
		return ginx.Result{Code: 500, Msg: "failed to get biz config"}, err
	}

	return ginx.Result{
		Code: 0,
		Msg:  "success",
		Data: gin.H{"config": h.toResponse(config)},
	}, nil
}

type UpdateBizConfigReq struct {
	ID        int64  `json:"id"`
	OwnerId   int64  `json:"owner_id"`
	OwnerType string `json:"owner_type"`
	Config    string `json:"config"`
}

func (h *BizConfigHandler) UpdateBizConfig(ctx *ginx.Context, req UpdateBizConfigReq, _ session.Session) (ginx.Result, error) {
	existing, err := h.svc.GetByID(ctx.Request.Context(), req.ID)
	if errors.Is(err, errs.ErrBizConfigNotFound) {
		return ginx.Result{Code: 404, Msg: "biz config not found"}, nil
	} else if err != nil {
		return ginx.Result{Code: 500, Msg: "failed to fetch biz config"}, err
	}

	// 更新字段
	existing.OwnerID = req.OwnerId
	existing.OwnerType = req.OwnerType
	existing.Config = req.Config

	if err := h.svc.Update(ctx.Request.Context(), existing); err != nil {
		return ginx.Result{Code: 500, Msg: "failed to update biz config"}, err
	}

	updated, err := h.svc.GetByID(ctx.Request.Context(), req.ID)
	if err != nil {
		return ginx.Result{Code: 500, Msg: "failed to fetch updated biz config"}, err
	}

	return ginx.Result{
		Code: 0,
		Msg:  "success",
		Data: gin.H{"config": h.toResponse(updated)},
	}, nil
}

type DeleteBizConfigReq struct {
	ID int64 `json:"id"`
}

func (h *BizConfigHandler) DeleteBizConfig(ctx *ginx.Context, req DeleteBizConfigReq, _ session.Session) (ginx.Result, error) {
	idStr := strconv.FormatInt(req.ID, 10)
	if err := h.svc.Delete(ctx.Request.Context(), idStr); err != nil {
		return ginx.Result{Code: 500, Msg: "failed to delete biz config"}, err
	}

	return ginx.Result{
		Code: 0,
		Msg:  "success",
		Data: gin.H{"success": true},
	}, nil
}

func (h *BizConfigHandler) toResponse(config domain.BizConfig) map[string]any {
	return map[string]any{
		"id":         config.ID,
		"owner_id":   config.OwnerID,
		"owner_type": config.OwnerType,
		"config":     config.Config,
		"ctime":      config.Ctime.Format("2006-01-02 15:04:05"),
		"utime":      config.Utime.Format("2006-01-02 15:04:05"),
	}
}

type ShardingExecution struct {
	// 独立的
	ExecID int64 // 任务 A 的执行 ID
}

type Execution struct {
	ParentID int64 // 任务 A 的执行 ID
}
