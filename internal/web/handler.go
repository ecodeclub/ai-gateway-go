package web

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.PromptService
}

func NewHandler(svc *service.PromptService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	prompt := server.Group("/prompt")
	prompt.POST("/add", ginx.BS(h.Add))
	prompt.GET("/:id", ginx.W(h.Get))
	prompt.POST("/delete", ginx.B(h.Delete))
	prompt.POST("/update", ginx.B(h.Update))
	prompt.POST("/publish", ginx.B(h.Publish))
	prompt.POST("/fork", ginx.B(h.Fork))
}

func (h *Handler) PublicRoutes(server *gin.Engine) {}

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
		Data: newGetVO(res),
	}, nil
}

// Delete 删除整个 prompt 或者某个版本
func (h *Handler) Delete(ctx *ginx.Context, req DeleteReq) (ginx.Result, error) {
	err := h.svc.Delete(ctx, req.ID, req.VersionID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// Update 更新 prompt 本身或者某个版本的信息
func (h *Handler) Update(ctx *ginx.Context, req UpdateReq) (ginx.Result, error) {
	prompt := domain.Prompt{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
	}
	if req.VersionID > 0 {
		version := domain.PromptVersion{
			ID:            req.VersionID,
			Content:       req.Content,
			SystemContent: req.SystemContent,
			Temperature:   req.Temperature,
			TopN:          req.TopN,
			MaxTokens:     req.MaxTokens,
		}
		prompt.Versions = []domain.PromptVersion{version}
	}
	err := h.svc.Update(ctx, prompt)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *Handler) Publish(ctx *ginx.Context, req PublishReq) (ginx.Result, error) {
	err := h.svc.Publish(ctx, req.ID, req.VersionID, req.Label)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// Fork 新增一个版本
func (h *Handler) Fork(ctx *ginx.Context, req ForkReq) (ginx.Result, error) {
	err := h.svc.Fork(ctx, req.ID, domain.PromptVersion{
		Content:       req.Content,
		SystemContent: req.SystemContent,
		Temperature:   req.Temperature,
		TopN:          req.TopN,
		MaxTokens:     req.MaxTokens,
	})
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}
