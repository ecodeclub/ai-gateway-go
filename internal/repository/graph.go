package repository

import (
	"context"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
)

// NodeRepo 是图结构的仓库实现，负责在领域模型和数据访问层之间进行转换。
type NodeRepo struct {
	graph *dao.GraphDAO // 数据访问对象，用于操作数据库
}

// NewGraphRepo 创建一个新的NodeRepo实例。
// 参数:
//
//	graph: 数据访问对象实例
//
// 返回值:
//
//	*NodeRepo: 初始化后的NodeRepo实例
func NewGraphRepo(graph *dao.GraphDAO) *NodeRepo {
	return &NodeRepo{graph: graph}
}

// SaveGraph 保存图结构到数据库。
// 如果图已经存在则更新它，否则创建新图。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	graph: 要保存的图结构
//
// 返回值:
//
//	int64: 保存后的图ID
//	error: 执行过程中发生的错误
func (n *NodeRepo) SaveGraph(ctx context.Context, graph domain.Graph) (int64, error) {
	str, _ := graph.Metadata.AsString()

	return n.graph.SaveGraph(ctx, dao.Graph{
		ID:       graph.ID,
		Metadata: str,
	})
}

// SaveNode 保存节点到数据库。
// 如果节点已经存在则更新它，否则创建新节点。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	node: 要保存的节点信息
//
// 返回值:
//
//	int64: 保存后的节点ID
//	error: 执行过程中发生的错误
func (n *NodeRepo) SaveNode(ctx context.Context, node domain.Node) (int64, error) {
	str, _ := node.Metadata.AsString()
	return n.graph.SaveNode(ctx, dao.Node{
		ID:       node.ID,
		Type:     node.Type,
		GraphID:  node.GraphID,
		Status:   node.Status,
		Metadata: str,
	})
}

// SaveEdge 保存边（连接）到数据库。
// 如果边已经存在则更新它，否则创建新边。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	edge: 要保存的边信息
//
// 返回值:
//
//	int64: 保存后的边ID
//	error: 执行过程中发生的错误
func (n *NodeRepo) SaveEdge(ctx context.Context, edge domain.Edge) (int64, error) {
	str, _ := edge.Metadata.AsString()

	return n.graph.SaveEdge(ctx, dao.Edge{
		ID:       edge.ID,
		GraphID:  edge.GraphID,
		SourceID: edge.SourceID,
		TargetID: edge.TargetID,
		Metadata: str,
	})
}

// DeleteGraph 根据ID删除图及其所有节点和边。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 要删除的图ID
//
// 返回值:
//
//	error: 执行过程中发生的错误
func (n *NodeRepo) DeleteGraph(ctx context.Context, id int64) error {
	return n.graph.DeleteGraph(ctx, id)
}

// DeleteNode 根据ID删除节点。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 要删除的节点ID
//
// 返回值:
//
//	error: 执行过程中发生的错误
func (n *NodeRepo) DeleteNode(ctx context.Context, id int64) error {
	return n.graph.DeleteNode(ctx, id)
}

// DeleteEdge 根据ID删除边。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 要删除的边ID
//
// 返回值:
//
//	error: 执行过程中发生的错误
func (n *NodeRepo) DeleteEdge(ctx context.Context, id int64) error {
	return n.graph.DeleteEdge(ctx, id)
}

// GetGraph 根据ID获取完整的图结构（包括所有节点和边）。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 要查询的图ID
//
// 返回值:
//
//	domain.Graph: 查询到的图结构
//	error: 执行过程中发生的错误
func (n *NodeRepo) GetGraph(ctx context.Context, id int64) (domain.Graph, error) {
	graph, err := n.graph.GetGraph(ctx, id)
	if err != nil {
		return domain.Graph{}, err
	}
	edges, err := n.graph.GetEdges(ctx, id)
	if err != nil {
		return domain.Graph{}, err
	}
	nodes, err := n.graph.GetNodes(ctx, id)
	if err != nil {
		return domain.Graph{}, err
	}
	return n.daoToDomain(graph, edges, nodes), nil
}

// daoToDomain 将DAO层的数据转换为领域模型的图结构。
// 参数:
//
//	graph: DAO层的图结构
//	edges: DAO层的边列表
//	nodes: DAO层的节点列表
//
// 返回值:
//
//	domain.Graph: 领域模型的图结构
func (n *NodeRepo) daoToDomain(graph dao.Graph, edges []dao.Edge, nodes []dao.Node) domain.Graph {
	var plan domain.Graph
	plan.ID = graph.ID
	steps := slice.Map[dao.Node, domain.Node](nodes, func(idx int, src dao.Node) domain.Node {
		metaData := src.Metadata
		return domain.Node{
			GraphID:  src.GraphID,
			ID:       src.ID,
			Type:     src.Type,
			Status:   src.Status,
			Metadata: ekit.AnyValue{Val: metaData},
		}
	})
	plan.Steps = steps

	domainEdges := slice.Map[dao.Edge, domain.Edge](edges, func(idx int, src dao.Edge) domain.Edge {
		metaData := src.Metadata
		return domain.Edge{
			GraphID:  src.GraphID,
			ID:       src.ID,
			SourceID: src.SourceID,
			TargetID: src.TargetID,
			Metadata: ekit.AnyValue{Val: metaData},
		}
	})

	plan.Edges = domainEdges
	return plan
}
