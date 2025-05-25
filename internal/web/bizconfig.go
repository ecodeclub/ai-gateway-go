package web

import (
	"strconv"

	"github.com/ecodeclub/ai-gateway-go/errs"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

// BizConfigHandler 处理与业务配置相关的 HTTP 请求
// 提供创建、获取、更新和删除业务配置的功能
type BizConfigHandler struct {
	svc service.BizConfigService // 业务逻辑层接口
}

// NewBizConfigHandler 创建一个新的 BizConfigHandler 实例
func NewBizConfigHandler(svc service.BizConfigService) *BizConfigHandler {
	return &BizConfigHandler{svc: svc}
}

// RegisterRoutes 注册业务配置相关的路由
func (h *BizConfigHandler) RegisterRoutes(server *gin.Engine) {
	bg := server.Group("/api/v1/biz-configs")

	bg.POST("/create", ginx.BS[CreateBizConfigReq](h.CreateBizConfig))
	bg.POST("/get", ginx.BS[GetBizConfigReq](h.GetBizConfig))
	bg.POST("/update", ginx.BS[UpdateBizConfigReq](h.UpdateBizConfig))
	bg.POST("/delete", ginx.BS[DeleteBizConfigReq](h.DeleteBizConfig))
}

// CreateBizConfigReq 定义创建业务配置的请求结构体
type CreateBizConfigReq struct {
	ID        int64  `json:"id"`
	OwnerId   int64  `json:"owner_id"`
	OwnerType string `json:"owner_type"`
	Config    string `json:"config"`
}

// CreateBizConfig 处理创建业务配置的 HTTP 请求
func (h *BizConfigHandler) CreateBizConfig(ctx *ginx.Context, req CreateBizConfigReq, sess session.Session) (ginx.Result, error) {
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

// GetBizConfigReq 定义获取业务配置的请求结构体
type GetBizConfigReq struct {
	ID int64 `json:"id"`
}

// GetBizConfig 处理获取业务配置的 HTTP 请求
func (h *BizConfigHandler) GetBizConfig(ctx *ginx.Context, req GetBizConfigReq, sess session.Session) (ginx.Result, error) {
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

// UpdateBizConfigReq 定义更新业务配置的请求结构体
type UpdateBizConfigReq struct {
	ID        int64  `json:"id"`
	OwnerId   int64  `json:"owner_id"`
	OwnerType string `json:"owner_type"`
	Config    string `json:"config"`
}

// UpdateBizConfig 处理更新业务配置的 HTTP 请求
func (h *BizConfigHandler) UpdateBizConfig(ctx *ginx.Context, req UpdateBizConfigReq, sess session.Session) (ginx.Result, error) {
	existing, err := h.svc.GetByID(ctx.Request.Context(), req.ID)
	if err == errs.ErrBizConfigNotFound {
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

// DeleteBizConfigReq 定义删除业务配置的请求结构体
type DeleteBizConfigReq struct {
	ID int64 `json:"id"`
}

// DeleteBizConfig 处理删除业务配置的 HTTP 请求
func (h *BizConfigHandler) DeleteBizConfig(ctx *ginx.Context, req DeleteBizConfigReq, sess session.Session) (ginx.Result, error) {
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

// toResponse 将领域模型转换为响应格式
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
