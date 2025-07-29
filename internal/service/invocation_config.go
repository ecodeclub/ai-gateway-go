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

func NewInvocationConfigService(repo *repository.InvocationConfigRepo,
	bizRepo *repository.BizConfigRepository, providerRepo *repository.ProviderRepo) *InvocationConfigService {
	return &InvocationConfigService{
		repo:         repo,
		bizRepo:      bizRepo,
		providerRepo: providerRepo}
}

func (s *InvocationConfigService) Save(ctx context.Context, cfg domain.InvocationConfig) (int64, error) {
	return s.repo.Save(ctx, cfg)
}

func (s *InvocationConfigService) SaveVersion(ctx context.Context, version domain.InvocationCfgVersion) (int64, error) {
	version.Status = domain.InvocationCfgVersionStatusDraft
	return s.repo.SaveVersion(ctx, version)
}

func (s *InvocationConfigService) Get(ctx context.Context, id int64) (domain.InvocationConfig, error) {
	var (
		eg  errgroup.Group
		res domain.InvocationConfig
		biz domain.BizConfig
	)
	eg.Go(func() error {
		var err error
		res, err = s.repo.Get(ctx, id)
		return err
	})
	eg.Go(func() error {
		var err error
		biz, err = s.bizRepo.GetByID(ctx, id)
		return err
	})
	err := eg.Wait()
	res.Biz = biz
	return res, err
}

func (s *InvocationConfigService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *InvocationConfigService) DeleteVersion(ctx context.Context, versionID int64) error {
	return s.repo.DeleteVersion(ctx, versionID)
}

func (s *InvocationConfigService) UpdateInfo(ctx context.Context, prompt domain.InvocationConfig) error {
	return s.repo.UpdateInfo(ctx, prompt)
}

func (s *InvocationConfigService) UpdateVersion(ctx context.Context, version domain.InvocationCfgVersion) error {
	return s.repo.UpdateVersion(ctx, version)
}

func (s *InvocationConfigService) Publish(ctx context.Context, versionID int64, label string) error {
	return s.repo.UpdateActiveVersion(ctx, versionID, label)
}

func (s *InvocationConfigService) Fork(ctx context.Context, versionID int64) (int64, error) {
	prompt, err := s.repo.GetByVersionID(ctx, versionID)
	if err != nil {
		return 0, err
	}
	prompt.ID = 0
	return s.repo.SaveVersion(ctx, prompt)
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

func (s *InvocationConfigService) ListVersions(ctx context.Context, invID int64, offset int, limit int) ([]domain.InvocationCfgVersion, int, error) {
	var (
		eg       errgroup.Group
		versions []domain.InvocationCfgVersion
		total    int
	)
	eg.Go(func() error {
		var err error
		versions, err = s.repo.GetVersions(ctx, invID, offset, limit)
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

func (s *InvocationConfigService) GetVersion(ctx context.Context, versionID int64) (domain.InvocationCfgVersion, error) {
	res, err := s.repo.GetByVersionID(ctx, versionID)
	if err != nil {
		return domain.InvocationCfgVersion{}, err
	}
	model, err := s.providerRepo.GetModel(ctx, res.Model.ID)
	if err != nil {
		return domain.InvocationCfgVersion{}, err
	}
	res.Model = model
	return res, nil
}

func (s *InvocationConfigService) ActivateVersion(ctx context.Context, id int64) error {
	return s.repo.ActivateVersion(ctx, id)
}
