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

package admin

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

type GraphHandler struct {
	svc *service.NodeService
}

func NewGraphHandler(nodeSvc *service.NodeService) *GraphHandler {
	return &GraphHandler{svc: nodeSvc}
}

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

func (h *GraphHandler) PublicRoutes(_ *gin.Engine) {
}

func (h *GraphHandler) GetGraph(ctx *ginx.Context, req GetReq, _ session.Session) (ginx.Result, error) {
	graph, err := h.svc.GetGraph(ctx, req.ID)
	if err != nil {
		elog.Error("获取graph 失败", elog.Int64("ID", req.ID), elog.Any("err", err))
		return ginx.Result{Code: 500, Msg: "内部错误"}, ginx.ErrNoResponse
	}
	return ginx.Result{Msg: "OK", Data: newGetNodeVO(graph)}, err
}

func (h *GraphHandler) SaveGraph(ctx *ginx.Context, req SaveGraphReq, _ session.Session) (ginx.Result, error) {
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

func (h *GraphHandler) SaveNode(ctx *ginx.Context, req Node, _ session.Session) (ginx.Result, error) {
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

func (h *GraphHandler) SaveEdge(ctx *ginx.Context, req Edge, _ session.Session) (ginx.Result, error) {
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

func (h *GraphHandler) DeleteNode(ctx *ginx.Context, req DeleteReq, _ session.Session) (ginx.Result, error) {
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

func (h *GraphHandler) DeleteEdge(ctx *ginx.Context, req DeleteReq, _ session.Session) (ginx.Result, error) {
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

func (h *GraphHandler) DeleteGraph(ctx *ginx.Context, req DeleteReq, _ session.Session) (ginx.Result, error) {
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
