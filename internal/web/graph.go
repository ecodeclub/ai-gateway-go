package web

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx"
	"github.com/gin-gonic/gin"
)

type NodeHandler struct {
	svc *service.NodeService
}

func NewNodeHandler(nodeSvc *service.NodeService) *NodeHandler {
	return &NodeHandler{svc: nodeSvc}
}

func (h *NodeHandler) PrivateRoute(engine *gin.Engine) {
	group := engine.Group("/node")
	group.GET("/get/:id", ginx.W(h.Get))
	group.POST("/save", ginx.B(h.Save))
}

func (h *NodeHandler) Get(ctx *ginx.Context) (ginx.Result, error) {
	id, err := ctx.Param("id").AsInt64()
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	plan, err := h.svc.Get(ctx, id)
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	return ginx.Result{Msg: "OK", Data: newGetNodeVO(plan)}, err
}

func (h *NodeHandler) Save(ctx *ginx.Context, req SavePlanReq) (ginx.Result, error) {
	var plan domain.Plan
	plan.ID = req.ID
	plan.Steps = slice.Map[Step, domain.Step](req.Steps, func(idx int, src Step) domain.Step {
		return domain.Step{ID: src.ID, Status: src.Status, Metadata: ekit.AnyValue{Val: src.Metadata}}
	})

	id, err := h.svc.Save(ctx, plan)
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	return ginx.Result{Msg: "OK", Data: id}, err
}
