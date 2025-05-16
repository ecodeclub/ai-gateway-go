package repository

import (
	"context"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
)

type NodeRepo struct {
	graph *dao.GraphDAO
}

func NewGraphRepo(graph *dao.GraphDAO) *NodeRepo {
	return &NodeRepo{graph: graph}
}

func (n *NodeRepo) SaveGraph(ctx context.Context, graph domain.Graph) (int64, error) {
	str, _ := graph.Metadata.AsString()

	return n.graph.SaveGraph(ctx, dao.Graph{
		ID:       graph.ID,
		Metadata: str,
	})
}

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

func (n *NodeRepo) DeleteGraph(ctx context.Context, id int64) error {
	return n.graph.DeleteGraph(ctx, id)
}

func (n *NodeRepo) DeleteNode(ctx context.Context, id int64) error {
	return n.graph.DeleteNode(ctx, id)
}

func (n *NodeRepo) DeleteEdge(ctx context.Context, id int64) error {
	return n.graph.DeleteEdge(ctx, id)
}

func (n *NodeRepo) GetGraph(ctx context.Context, id int64) (domain.Graph, error) {
	graph, err := n.graph.Get(ctx, id)
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
