package service

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

type NodeService struct {
	repo *repository.NodeRepo
}

func NewNodeService(repo *repository.NodeRepo) *NodeService {
	return &NodeService{repo: repo}
}

func (svc *NodeService) Get(ctx context.Context, id int64) (domain.Plan, error) {
	return svc.repo.Get(ctx, id)
}

func (svc *NodeService) Save(ctx context.Context, graph domain.Plan) (int64, error) {
	return svc.repo.Save(ctx, graph)
}
