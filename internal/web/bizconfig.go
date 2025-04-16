package web

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type BizConfigHandler struct {
	svc *service.BizConfigService
}

func NewBizConfigHandler(svc *service.BizConfigService) *BizConfigHandler {
	return &BizConfigHandler{svc: svc}
}

func (h *BizConfigHandler) RegisterRoutes(server *gin.Engine) {
	bg := server.Group("/api/v1/biz-configs")

	bg.POST("", h.CreateBizConfig)       // 创建
	bg.GET("/:id", h.GetBizConfig)       // 查询单个
	bg.PUT("/:id", h.UpdateBizConfig)    // 更新
	bg.DELETE("/:id", h.DeleteBizConfig) // 删除
}

func (h *BizConfigHandler) CreateBizConfig(c *gin.Context) {
	var CreateBizConfigReq struct {
		ID        int64  `json:"id"`
		OwnerId   int64  `json:"owner_id"`
		OwnerType string `json:"owner_type"`
		Config    string `json:"config"`
		Token     string `json:"token"`
	}
	if err := c.ShouldBindJSON(&CreateBizConfigReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	config := domain.BizConfig{
		ID:        CreateBizConfigReq.ID,
		OwnerID:   CreateBizConfigReq.OwnerId,
		OwnerType: CreateBizConfigReq.OwnerType,
		Config:    CreateBizConfigReq.Config,
		Token:     CreateBizConfigReq.Token,
	}

	created, token, err := h.svc.Create(c.Request.Context(), config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create biz config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bizconfig":    h.toResponse(created),
		"access_token": token,
	})
}

func (h *BizConfigHandler) GetBizConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	config, err := h.svc.GetByID(c.Request.Context(), id)
	if err == service.ErrBizConfigNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "biz config not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get biz config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": h.toResponse(config)})
}

func (h *BizConfigHandler) UpdateBizConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var UpdateBizConfigReq struct {
		Config string `json:"config"`
	}
	if err := c.ShouldBindJSON(&UpdateBizConfigReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	existing, err := h.svc.GetByID(c.Request.Context(), id)
	if err == service.ErrBizConfigNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "biz config not found"})
		return
	}

	existing.Config = UpdateBizConfigReq.Config

	if err := h.svc.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update biz config"})
		return
	}

	updated, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated biz config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": h.toResponse(updated)})
}

func (h *BizConfigHandler) DeleteBizConfig(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete biz config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BizConfigHandler) toResponse(config domain.BizConfig) map[string]any {
	return map[string]any{
		"id":         config.ID,
		"owner_id":   config.OwnerID,
		"owner_type": config.OwnerType,
		"token":      config.Token,
		"config":     config.Config,
		"created_at": config.CreatedAt,
		"updated_at": config.UpdatedAt,
	}
}
