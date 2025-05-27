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

package service

import (
	"context"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

type NodeService struct {
	repo *repository.NodeRepo
}

func NewGraphService(repo *repository.NodeRepo) *NodeService {
	return &NodeService{repo: repo}
}

func (svc *NodeService) GetGraph(ctx context.Context, id int64) (domain.Graph, error) {
	return svc.repo.GetGraph(ctx, id)
}

func (svc *NodeService) SaveGraph(ctx context.Context, graph domain.Graph) (int64, error) {
	return svc.repo.SaveGraph(ctx, graph)
}

func (svc *NodeService) SaveNode(ctx context.Context, step domain.Node) (int64, error) {
	return svc.repo.SaveNode(ctx, step)
}

func (svc *NodeService) SaveEdge(ctx context.Context, edge domain.Edge) (int64, error) {
	return svc.repo.SaveEdge(ctx, edge)
}

func (svc *NodeService) DeleteEdge(ctx context.Context, id int64) error {
	return svc.repo.DeleteEdge(ctx, id)
}

func (svc *NodeService) DeleteNode(ctx context.Context, id int64) error {
	return svc.repo.DeleteNode(ctx, id)
}

func (svc *NodeService) DeleteGraph(ctx context.Context, id int64) error {
	return svc.repo.DeleteGraph(ctx, id)
}
