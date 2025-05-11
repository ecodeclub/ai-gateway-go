package repository

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
)

type NodeRepo struct {
	graph *dao.GraphDao
}

func NewNodeRepo(graph *dao.GraphDao) *NodeRepo {
	return &NodeRepo{graph: graph}
}

func (n *NodeRepo) Save(ctx context.Context, graph domain.Plan) (int64, error) {
	g := n.domainToDao(graph)
	return n.graph.Save(ctx, g)
}

func (n *NodeRepo) Get(ctx context.Context, id int64) (domain.Plan, error) {
	graph, err := n.graph.Get(ctx, id)
	if err != nil {
		return domain.Plan{}, err
	}
	return n.daoToDomain(graph), nil
}

func (n *NodeRepo) domainToDao(plan domain.Plan) dao.Graph {
	var g dao.Graph
	g.ID = plan.ID

	nodes := slice.Map[domain.Step, dao.Node](plan.Steps, func(idx int, src domain.Step) dao.Node {
		metaData, _ := src.Metadata.AsString()
		return dao.Node{
			ID:       src.ID,
			Type:     src.Type,
			Status:   src.Status,
			Metadata: metaData,
		}
	})
	g.Nodes = nodes
	return g
}

func (n *NodeRepo) daoToDomain(graph dao.Graph) domain.Plan {
	var plan domain.Plan
	plan.ID = graph.ID
	nodes := slice.Map[dao.Node, domain.Step](graph.Nodes, func(idx int, src dao.Node) domain.Step {
		metaData := src.Metadata
		return domain.Step{
			ID:       src.ID,
			Type:     src.Type,
			Status:   src.Status,
			Metadata: ekit.AnyValue{Val: metaData},
		}
	})
	plan.Steps = nodes
	return plan
}
