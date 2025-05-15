package web

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

type GraphHandler struct {
	svc *service.NodeService
}

func NewGraphHandler(nodeSvc *service.NodeService) *GraphHandler {
	return &GraphHandler{svc: nodeSvc}
}

func (h *GraphHandler) PrivateRoutes(engine *gin.Engine) {
	graph := engine.Group("/graph")
	{
		graph.POST("/save", ginx.BS[SaveGraphReq](h.SaveGraph))
		graph.POST("/delete", ginx.BS[DeleteReq](h.DeleteGraph))
		graph.POST("/get", ginx.BS[GetReq](h.GetGraph))
	}

	node := engine.Group("/node")
	{
		node.POST("/save", ginx.BS[Node](h.SaveNode))
		node.POST("/delete", ginx.BS[DeleteReq](h.DeleteNode))
	}

	edge := engine.Group("/edge")
	{
		edge.POST("/save", ginx.BS[Edge](h.SaveEdge))
		edge.POST("/delete", ginx.BS[DeleteReq](h.DeleteEdge))
	}
}

func (h *GraphHandler) PublicRoutes(engine *gin.Engine) {

}

func (h *GraphHandler) GetGraph(ctx *ginx.Context, req GetReq, sess session.Session) (ginx.Result, error) {
	graph, err := h.svc.GetGraph(ctx, req.ID)
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	return ginx.Result{Msg: "OK", Data: newGetNodeVO(graph)}, err
}

func (h *GraphHandler) SaveGraph(ctx *ginx.Context, req SaveGraphReq, sess session.Session) (ginx.Result, error) {
	graph := domain.Graph{
		ID: req.ID,
	}

	id, err := h.svc.SaveGraph(ctx, graph)
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}

	return ginx.Result{Msg: "OK", Data: id}, nil
}

func (h *GraphHandler) SaveNode(ctx *ginx.Context, req Node, sess session.Session) (ginx.Result, error) {
	node := domain.Node{
		ID:       req.ID,
		GraphID:  req.GraphID,
		Type:     req.Type,
		Status:   req.Status,
		Metadata: ekit.AnyValue{Val: req.Metadata},
	}

	id, err := h.svc.SaveNode(ctx, node)
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	return ginx.Result{Data: id}, nil
}

func (h *GraphHandler) SaveEdge(ctx *ginx.Context, req Edge, sess session.Session) (ginx.Result, error) {
	edge := domain.Edge{
		ID:       req.ID,
		GraphID:  req.GraphID,
		TargetID: req.TargetID,
		SourceID: req.SourceID,
		Metadata: ekit.AnyValue{Val: req.Metadata},
	}

	id, err := h.svc.SaveEdge(ctx, edge)
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}

	return ginx.Result{
		Data: id,
	}, nil
}

func (h *GraphHandler) DeleteNode(ctx *ginx.Context, req DeleteReq, sess session.Session) (ginx.Result, error) {
	id := req.ID

	err := h.svc.DeleteNode(ctx, id)
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *GraphHandler) DeleteEdge(ctx *ginx.Context, req DeleteReq, sess session.Session) (ginx.Result, error) {
	id := req.ID
	err := h.svc.DeleteEdge(ctx, id)
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *GraphHandler) DeleteGraph(ctx *ginx.Context, req DeleteReq, sess session.Session) (ginx.Result, error) {
	id := req.ID
	err := h.svc.DeleteGraph(ctx, id)
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}
