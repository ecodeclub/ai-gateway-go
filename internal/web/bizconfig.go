package web

import (
	"github.com/ecodeclub/ai-gateway-go/errs"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ginx"
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

	bg.POST("/create", h.CreateBizConfig)
	bg.POST("/get", h.GetBizConfig)
	bg.POST("/update", h.UpdateBizConfig)
	bg.POST("/delete", h.DeleteBizConfig)
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
		c.JSON(http.StatusBadRequest, ginx.Result{Code: 400, Msg: "invalid request"})
		return
	}

	config := domain.BizConfig{
		ID:        CreateBizConfigReq.ID,
		OwnerID:   CreateBizConfigReq.OwnerId,
		OwnerType: CreateBizConfigReq.OwnerType,
		Config:    CreateBizConfigReq.Config,
		Token:     CreateBizConfigReq.Token,
	}

	created, access_token, err := h.svc.Create(c.Request.Context(), config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginx.Result{Code: 500, Msg: "failed to create biz config"})
		return
	}

	c.JSON(http.StatusOK, ginx.Result{
		Code: 0,
		Msg:  "success",
		Data: gin.H{
			"bizconfig":    h.toResponse(created),
			"access_token": access_token,
		},
	})
}

func (h *BizConfigHandler) GetBizConfig(c *gin.Context) {
	var GetBizConfigReq struct {
		ID int64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&GetBizConfigReq); err != nil {
		c.JSON(http.StatusBadRequest, ginx.Result{Code: 400, Msg: "invalid request"})
		return
	}

	config, err := h.svc.GetByID(c.Request.Context(), GetBizConfigReq.ID)
	if err == errs.ErrBizConfigNotFound {
		c.JSON(http.StatusNotFound, ginx.Result{Code: 404, Msg: "biz config not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, ginx.Result{Code: 500, Msg: "failed to get biz config"})
		return
	}

	c.JSON(http.StatusOK, ginx.Result{
		Code: 0,
		Msg:  "success",
		Data: gin.H{"config": h.toResponse(config)},
	})
}

func (h *BizConfigHandler) UpdateBizConfig(c *gin.Context) {
	var UpdateBizConfigReq struct {
		ID        int64  `json:"id"`
		OwnerId   int64  `json:"owner_id"`
		OwnerType string `json:"owner_type"`
		Config    string `json:"config"`
		Token     string `json:"token"`
	}
	if err := c.ShouldBindJSON(&UpdateBizConfigReq); err != nil {
		c.JSON(http.StatusBadRequest, ginx.Result{Code: 400, Msg: "invalid request"})
		return
	}

	existing, err := h.svc.GetByID(c.Request.Context(), UpdateBizConfigReq.ID)
	if err == errs.ErrBizConfigNotFound {
		c.JSON(http.StatusNotFound, ginx.Result{Code: 404, Msg: "biz config not found"})
		return
	}

	existing.OwnerID = UpdateBizConfigReq.OwnerId
	existing.OwnerType = UpdateBizConfigReq.OwnerType
	existing.Config = UpdateBizConfigReq.Config
	existing.Token = UpdateBizConfigReq.Token

	if err := h.svc.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, ginx.Result{Code: 500, Msg: "failed to update biz config"})
		return
	}

	updated, err := h.svc.GetByID(c.Request.Context(), UpdateBizConfigReq.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginx.Result{Code: 500, Msg: "failed to fetch updated biz config"})
		return
	}

	c.JSON(http.StatusOK, ginx.Result{
		Code: 0,
		Msg:  "success",
		Data: gin.H{"config": h.toResponse(updated)},
	})
}

func (h *BizConfigHandler) DeleteBizConfig(c *gin.Context) {
	var DeleteBizConfigReq struct {
		ID int64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&DeleteBizConfigReq); err != nil {
		c.JSON(http.StatusBadRequest, ginx.Result{Code: 400, Msg: "invalid request"})
		return
	}

	idStr := strconv.FormatInt(DeleteBizConfigReq.ID, 10)
	if err := h.svc.Delete(c.Request.Context(), idStr); err != nil {
		c.JSON(http.StatusInternalServerError, ginx.Result{Code: 500, Msg: "failed to delete biz config"})
		return
	}

	c.JSON(http.StatusOK, ginx.Result{
		Code: 0,
		Msg:  "success",
		Data: gin.H{"success": true},
	})
}

func (h *BizConfigHandler) toResponse(config domain.BizConfig) map[string]any {
	return map[string]any{
		"id":         config.ID,
		"owner_id":   config.OwnerID,
		"owner_type": config.OwnerType,
		"token":      config.Token,
		"config":     config.Config,
		"ctime":      config.Ctime.Format("2006-01-02 15:04:05"),
		"utime":      config.Utime.Format("2006-01-02 15:04:05"),
	}
}
