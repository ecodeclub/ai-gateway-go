package web

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
)

// GraphHandler 处理与图相关的 HTTP 请求
// 提供保存、删除、查询图信息的功能
type GraphHandler struct {
	svc *service.NodeService // 图相关业务逻辑接口
}

// NewGraphHandler 创建一个新的 GraphHandler 实例
func NewGraphHandler(nodeSvc *service.NodeService) *GraphHandler {
	return &GraphHandler{svc: nodeSvc}
}

// PrivateRoutes 注册私有路由（需要身份验证）
func (h *GraphHandler) PrivateRoutes(engine *gin.Engine) {
	graph := engine.Group("/graph")
	graph.POST("/save", ginx.BS[SaveGraphReq](h.SaveGraph))
	graph.POST("/delete", ginx.BS[DeleteReq](h.DeleteGraph))
	graph.POST("/detail", ginx.BS[GetReq](h.GetGraph))

	node := engine.Group("/node")
	node.POST("/save", ginx.BS[Node](h.SaveNode))
	node.POST("/delete", ginx.BS[DeleteReq](h.DeleteNode))

	edge := engine.Group("/edge")
	edge.POST("/save", ginx.BS[Edge](h.SaveEdge))
	edge.POST("/delete", ginx.BS[DeleteReq](h.DeleteEdge))
}

// PublicRoutes 注册公共路由（不需要身份验证）
func (h *GraphHandler) PublicRoutes(engine *gin.Engine) {

}

// GetGraph 处理获取图信息的 HTTP 请求
func (h *GraphHandler) GetGraph(ctx *ginx.Context, req GetReq, sess session.Session) (ginx.Result, error) {
	graph, err := h.svc.GetGraph(ctx, req.ID)
	if err != nil {
		elog.Error("获取graph 失败", elog.Int64("ID", req.ID), elog.Any("err", err))
		return ginx.Result{Code: 500, Msg: "内部错误"}, ginx.ErrNoResponse
	}
	return ginx.Result{Msg: "OK", Data: newGetNodeVO(graph)}, err
}

// SaveGraph 处理保存图信息的 HTTP 请求
func (h *GraphHandler) SaveGraph(ctx *ginx.Context, req SaveGraphReq, sess session.Session) (ginx.Result, error) {
	graph := domain.Graph{
		ID: req.ID,
	}

	id, err := h.svc.SaveGraph(ctx, graph)
	if err != nil {
		elog.Error("保存 Graph 失败", elog.Int64("ID", req.ID), elog.Any("err", err))
		return ginx.Result{Code: 500, Msg: "内部错误"}, ginx.ErrNoResponse
	}

	return ginx.Result{Msg: "OK", Data: id}, nil
}

// SaveNode 处理保存节点信息的 HTTP 请求
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
		elog.Error("保存 Node 失败", elog.Int64("ID", req.ID), elog.Any("err", err))
		return ginx.Result{Code: 500, Msg: "内部错误"}, ginx.ErrNoResponse
	}
	return ginx.Result{Data: id}, nil
}

// SaveEdge 处理保存边信息的 HTTP 请求
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
		elog.Error("保存 Edge 失败", elog.Int64("ID", req.ID), elog.Any("err", err))
		return ginx.Result{Code: 500, Msg: "内部错误"}, ginx.ErrNoResponse
	}

	return ginx.Result{
		Data: id,
	}, nil
}

// DeleteNode 处理删除节点的 HTTP 请求
func (h *GraphHandler) DeleteNode(ctx *ginx.Context, req DeleteReq, sess session.Session) (ginx.Result, error) {
	id := req.ID

	err := h.svc.DeleteNode(ctx, id)
	if err != nil {
		elog.Error("删除 Node 失败", elog.Int64("ID", req.ID), elog.Any("err", err))
		return ginx.Result{Code: 500, Msg: "内部错误"}, ginx.ErrNoResponse
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// DeleteEdge 处理删除边的 HTTP 请求
func (h *GraphHandler) DeleteEdge(ctx *ginx.Context, req DeleteReq, sess session.Session) (ginx.Result, error) {
	id := req.ID
	err := h.svc.DeleteEdge(ctx, id)
	if err != nil {
		elog.Error("删除 Edge 失败", elog.Int64("ID", req.ID), zap.Any("err", err))
		return ginx.Result{Code: 500, Msg: "内部错误"}, ginx.ErrNoResponse
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// DeleteGraph 处理删除整个图的 HTTP 请求
func (h *GraphHandler) DeleteGraph(ctx *ginx.Context, req DeleteReq, sess session.Session) (ginx.Result, error) {
	id := req.ID
	err := h.svc.DeleteGraph(ctx, id)
	if err != nil {
		elog.Error("删除 Graph 失败", elog.Int64("ID", req.ID), elog.Any("err", err))
		return ginx.Result{Code: 500, Msg: "内部错误"}, ginx.ErrNoResponse
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}
