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

type BizConfigService interface {
	Create(ctx context.Context, config domain.BizConfig) (domain.BizConfig, error)
	GetByID(ctx context.Context, id int64) (domain.BizConfig, error)
	Update(ctx context.Context, config domain.BizConfig) error
	Delete(ctx context.Context, id string) error
}

type bizConfigService struct {
	repo *repository.BizConfigRepository
}

func NewBizConfigService(repo *repository.BizConfigRepository) BizConfigService {
	return &bizConfigService{
		repo: repo,
	}
}

func (s *bizConfigService) Create(ctx context.Context, req domain.BizConfig) (domain.BizConfig, error) {
	config := domain.BizConfig{
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Config:    req.Config,
	}

	created, err := s.repo.Create(ctx, config)
	if err != nil {
		return domain.BizConfig{}, err
	}

	return created, nil
}

func (s *bizConfigService) GetByID(ctx context.Context, id int64) (domain.BizConfig, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *bizConfigService) Update(ctx context.Context, config domain.BizConfig) error {
	return s.repo.Update(ctx, config)
}

func (s *bizConfigService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
