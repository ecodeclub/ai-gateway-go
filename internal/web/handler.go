package web

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

// Handler 处理与提示相关的 HTTP 请求
// 提供添加、获取、更新、删除提示及其版本管理的功能
type Handler struct {
	svc *service.PromptService // 提示相关业务逻辑接口
}

// NewHandler 创建一个新的 Handler 实例
func NewHandler(svc *service.PromptService) *Handler {
	return &Handler{svc: svc}
}

// PrivateRoutes 注册私有路由（需要身份验证）
func (h *Handler) PrivateRoutes(server *gin.Engine) {
	prompt := server.Group("/prompt")
	prompt.POST("/add", ginx.BS(h.Add))
	prompt.GET("/:id", ginx.W(h.Get))
	prompt.POST("/delete", ginx.B(h.Delete))
	prompt.POST("/delete/version", ginx.B(h.DeleteVersion))
	prompt.POST("/update", ginx.B(h.UpdatePrompt))
	prompt.POST("/update/version", ginx.B(h.UpdateVersion))
	prompt.POST("/publish", ginx.B(h.Publish))
	prompt.POST("/fork", ginx.B(h.Fork))
}

// PublicRoutes 注册公共路由（不需要身份验证）
func (h *Handler) PublicRoutes(server *gin.Engine) {}

// Add 处理添加新提示及其初始版本的 HTTP 请求
func (h *Handler) Add(ctx *ginx.Context, req AddReq, sess session.Session) (ginx.Result, error) {
	uid := sess.Claims().Uid
	// 这里我假设 owner_type 也存储在 jwt token 里
	ownerType, err := sess.Claims().Get("owner_type").String()
	if err != nil {
		return ginx.Result{}, ginx.ErrUnauthorized
	}
	prompt := domain.Prompt{
		Name:        req.Name,
		Description: req.Description,
		Owner:       uid,
		OwnerType:   domain.OwnerType(ownerType),
	}
	version := domain.PromptVersion{
		Content:       req.Content,
		SystemContent: req.SystemContent,
		Temperature:   req.Temperature,
		TopN:          req.TopN,
		MaxTokens:     req.MaxTokens,
	}
	err = h.svc.Add(ctx, prompt, version)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// Get 处理获取提示信息的 HTTP 请求
func (h *Handler) Get(ctx *ginx.Context) (ginx.Result, error) {
	id, err := ctx.Param("id").AsInt64()
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	res, err := h.svc.Get(ctx, id)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: newPromptVO(res),
	}, nil
}

// Delete 处理删除整个提示（软删除）的 HTTP 请求
func (h *Handler) Delete(ctx *ginx.Context, req DeleteReq) (ginx.Result, error) {
	err := h.svc.Delete(ctx, req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// DeleteVersion 处理删除特定提示版本的 HTTP 请求
func (h *Handler) DeleteVersion(ctx *ginx.Context, req DeleteVersionReq) (ginx.Result, error) {
	err := h.svc.DeleteVersion(ctx, req.VersionID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// UpdatePrompt 处理更新提示基本信息的 HTTP 请求
func (h *Handler) UpdatePrompt(ctx *ginx.Context, req UpdatePromptReq) (ginx.Result, error) {
	prompt := domain.Prompt{
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

// UpdateVersion 处理更新提示版本信息的 HTTP 请求
func (h *Handler) UpdateVersion(ctx *ginx.Context, req UpdateVersionReq) (ginx.Result, error) {
	version := domain.PromptVersion{
		ID:            req.VersionID,
		Content:       req.Content,
		SystemContent: req.SystemContent,
		Temperature:   req.Temperature,
		TopN:          req.TopN,
		MaxTokens:     req.MaxTokens,
	}
	err := h.svc.UpdateVersion(ctx, version)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// Publish 处理发布特定提示版本的 HTTP 请求
func (h *Handler) Publish(ctx *ginx.Context, req PublishReq) (ginx.Result, error) {
	err := h.svc.Publish(ctx, req.VersionID, req.Label)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// Fork 处理复制（fork）提示版本的 HTTP 请求
func (h *Handler) Fork(ctx *ginx.Context, req ForkReq) (ginx.Result, error) {
	err := h.svc.Fork(ctx, req.VersionID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}
