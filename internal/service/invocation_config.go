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

type InvocationConfigService struct {
	repo         *repository.InvocationConfigRepo
	bizRepo      *repository.BizConfigRepository
	providerRepo *repository.ProviderRepo
}

func NewInvocationConfigService(
	repo *repository.InvocationConfigRepo,
	bizRepo *repository.BizConfigRepository,
	providerRepo *repository.ProviderRepo,
) *InvocationConfigService {
	return &InvocationConfigService{
		repo:         repo,
		bizRepo:      bizRepo,
		providerRepo: providerRepo}
}

func (s *InvocationConfigService) Save(ctx context.Context, cfg domain.InvocationConfig) (int64, error) {
	return s.repo.Save(ctx, cfg)
}

func (s *InvocationConfigService) List(ctx context.Context, offset int, limit int) ([]domain.InvocationConfig, int, error) {
	var (
		eg    errgroup.Group
		cfgs  []domain.InvocationConfig
		total int
	)
	eg.Go(func() error {
		var err error
		cfgs, err = s.repo.List(ctx, offset, limit)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.Count(ctx)
		return err
	})
	return cfgs, total, eg.Wait()
}

func (s *InvocationConfigService) Detail(ctx context.Context, id int64) (domain.InvocationConfig, error) {
	res, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.InvocationConfig{}, err
	}
	biz, err := s.bizRepo.GetByID(ctx, res.Biz.ID)
	if err != nil {
		return domain.InvocationConfig{}, err
	}
	res.Biz = biz
	return res, nil
}

func (s *InvocationConfigService) SaveVersion(ctx context.Context, version domain.InvocationConfigVersion) (int64, error) {
	return s.repo.SaveVersion(ctx, version)
}

func (s *InvocationConfigService) ListVersions(ctx context.Context, invID int64, offset int, limit int) ([]domain.InvocationConfigVersion, int, error) {
	var (
		eg       errgroup.Group
		versions []domain.InvocationConfigVersion
		total    int
	)
	eg.Go(func() error {
		var err error
		versions, err = s.repo.ListVersions(ctx, invID, offset, limit)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.CountVersions(ctx, invID)
		return err
	})
	err := eg.Wait()
	return versions, total, err
}

func (s *InvocationConfigService) VersionDetail(ctx context.Context, versionID int64) (domain.InvocationConfigVersion, error) {
	var (
		eg     errgroup.Group
		res    domain.InvocationConfigVersion
		config domain.InvocationConfig
		model  domain.Model
	)

	res, err := s.repo.GetVersionByID(ctx, versionID)
	if err != nil {
		return domain.InvocationConfigVersion{}, err
	}

	eg.Go(func() error {
		var err1 error
		config, err1 = s.repo.Get(ctx, res.Config.ID)
		return err1
	})

	eg.Go(func() error {
		var err2 error
		model, err2 = s.providerRepo.GetModel(ctx, res.Model.ID)
		return err2
	})

	err = eg.Wait()
	res.Config = config
	res.Model = model
	return res, err
}

func (s *InvocationConfigService) ActivateVersion(ctx context.Context, id int64) error {
	return s.repo.ActivateVersion(ctx, id)
}

func (s *InvocationConfigService) ForkVersion(ctx context.Context, versionID int64) (int64, error) {
	version, err := s.repo.GetVersionByID(ctx, versionID)
	if err != nil {
		return 0, err
	}
	version.ID = 0
	version.Status = domain.InvocationCfgVersionStatusDraft
	return s.repo.SaveVersion(ctx, version)
}
