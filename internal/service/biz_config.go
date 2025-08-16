// Copyright 2025 ecodeclub
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
	"golang.org/x/sync/errgroup"
)

type BizConfigService struct {
	repo *repository.BizConfigRepository
}

func NewBizConfigService(repo *repository.BizConfigRepository) *BizConfigService {
	return &BizConfigService{
		repo: repo,
	}
}

func (s *BizConfigService) Save(ctx context.Context, req domain.BizConfig) (int64, error) {
	return s.repo.Save(ctx, req)
}

func (s *BizConfigService) List(ctx context.Context, offset, limit int) ([]domain.BizConfig, int64, error) {
	var (
		eg    errgroup.Group
		res   []domain.BizConfig
		total int64
	)
	eg.Go(func() error {
		var err error
		res, err = s.repo.List(ctx, offset, limit)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.Count(ctx)
		return err
	})
	err := eg.Wait()
	return res, total, err
}

func (s *BizConfigService) Detail(ctx context.Context, id int64) (domain.BizConfig, error) {
	return s.repo.GetByID(ctx, id)
}
