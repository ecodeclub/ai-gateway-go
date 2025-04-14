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
	prompt.DELETE("/:id", ginx.W(h.Delete))
	prompt.POST("/:id", ginx.B(h.Update))
}

func (h *Handler) PublicRoutes(server *gin.Engine) {}

func (h *Handler) Add(ctx *ginx.Context, req AddReq, sess session.Session) (ginx.Result, error) {
	uid := sess.Claims().Uid
	// 这里我假设 owner_type 也存储在 jwt token 里
	ownerType, err := sess.Claims().Get("owner_type").String()
	if err != nil {
		return ginx.Result{}, ginx.ErrUnauthorized
	}
	err = h.svc.Add(ctx, domain.Prompt{
		Name:        req.Name,
		Content:     req.Content,
		Description: req.Description,
		Owner:       uid,
		OwnerType:   domain.OwnerType(ownerType),
	})
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

func (h *Handler) Delete(ctx *ginx.Context) (ginx.Result, error) {
	id, err := ctx.Param("id").AsInt64()
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	err = h.svc.Delete(ctx, id)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *Handler) Update(ctx *ginx.Context, req UpdateReq) (ginx.Result, error) {
	id, err := ctx.Param("id").AsInt64()
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	err = h.svc.Update(ctx, id, req.Name, req.Content, req.Description)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}
