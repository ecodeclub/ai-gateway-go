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

package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
)

type InvocationConfigRepo struct {
	dao *dao.InvocationConfigDAO
}

func NewInvocationConfigRepo(dao *dao.InvocationConfigDAO) *InvocationConfigRepo {
	return &InvocationConfigRepo{dao: dao}
}

func (p *InvocationConfigRepo) Save(ctx context.Context, prompt domain.InvocationConfig) (int64, error) {
	return p.dao.Save(ctx, p.toDAO(prompt))
}

func (p *InvocationConfigRepo) Get(ctx context.Context, id int64) (domain.InvocationConfig, error) {
	cfg, err := p.dao.FindByID(ctx, id)
	if err != nil {
		return domain.InvocationConfig{}, err
	}
	return domain.InvocationConfig{
		ID:   cfg.ID,
		Name: cfg.Name,
		Biz: domain.BizConfig{
			ID: id,
		},
		Description: cfg.Description,
		Ctime:       time.UnixMilli(cfg.Ctime),
		Utime:       time.UnixMilli(cfg.Utime),
	}, err
}

func (p *InvocationConfigRepo) Delete(ctx context.Context, id int64) error {
	return p.dao.Delete(ctx, id)
}

func (p *InvocationConfigRepo) DeleteVersion(ctx context.Context, versionID int64) error {
	return p.dao.DeleteVersion(ctx, versionID)
}

func (p *InvocationConfigRepo) UpdateInfo(ctx context.Context, value domain.InvocationConfig) error {
	return p.dao.UpdatePrompt(ctx, dao.InvocationConfig{
		ID:          value.ID,
		Name:        value.Name,
		Description: value.Description,
	})
}

func (p *InvocationConfigRepo) UpdateVersion(ctx context.Context, value domain.InvocationCfgVersion) error {
	return p.dao.UpdateVersion(ctx, dao.InvocationConfigVersion{
		ID:           value.ID,
		Prompt:       value.Prompt,
		SystemPrompt: value.SystemPrompt,
		Temperature:  value.Temperature,
		TopP:         value.TopP,
		MaxTokens:    value.MaxTokens,
	})
}

func (p *InvocationConfigRepo) UpdateActiveVersion(ctx context.Context, versionID int64, label string) error {
	return p.dao.UpdateActiveVersion(ctx, versionID, label)
}

func (p *InvocationConfigRepo) InsertVersion(ctx context.Context, invID int64, version domain.InvocationCfgVersion) error {
	return p.dao.InsertVersion(ctx, dao.InvocationConfigVersion{
		InvID:        invID,
		Prompt:       version.Prompt,
		SystemPrompt: version.SystemPrompt,
		Temperature:  version.Temperature,
		TopP:         version.TopP,
		MaxTokens:    version.MaxTokens,
	})
}

func (p *InvocationConfigRepo) GetByVersionID(ctx context.Context, id int64) (domain.InvocationCfgVersion, error) {
	res, err := p.dao.GetByVersionID(ctx, id)
	if err != nil {
		return domain.InvocationCfgVersion{}, err
	}
	return p.toDomainVersion(res), nil
}

func (p *InvocationConfigRepo) toDomainVersion(v dao.InvocationConfigVersion) domain.InvocationCfgVersion {
	return domain.InvocationCfgVersion{
		ID:           v.ID,
		InvID:        v.InvID,
		Version:      v.Version,
		Model:        domain.Model{ID: v.ModelID},
		Prompt:       v.Prompt,
		SystemPrompt: v.SystemPrompt,
		Temperature:  v.Temperature,
		TopP:         v.TopP,
		MaxTokens:    v.MaxTokens,
		Status:       domain.InvocationCfgVersionStatus(v.Status),
		Ctime:        time.UnixMilli(v.Ctime),
		Utime:        time.UnixMilli(v.Utime),
	}
}

func (p *InvocationConfigRepo) toDAO(src domain.InvocationConfig) dao.InvocationConfig {
	return dao.InvocationConfig{
		ID:          src.ID,
		Name:        src.Name,
		BizID:       src.Biz.ID,
		Description: src.Description,
	}
}

func (p *InvocationConfigRepo) toDAOVersion(src domain.InvocationCfgVersion) dao.InvocationConfigVersion {
	return dao.InvocationConfigVersion{
		ID:           src.ID,
		InvID:        src.InvID,
		ModelID:      src.Model.ID,
		Prompt:       src.Prompt,
		SystemPrompt: src.SystemPrompt,
		Version:      src.Version,
		Temperature:  src.Temperature,
		TopP:         src.TopP,
		MaxTokens:    src.MaxTokens,
		Status:       src.Status.String(),
	}
}

func (p *InvocationConfigRepo) List(ctx context.Context, offset int, limit int) ([]domain.InvocationConfig, error) {
	res, err := p.dao.List(ctx, offset, limit)
	return slice.Map(res, func(idx int, src dao.InvocationConfig) domain.InvocationConfig {
		return p.toDomain(src)
	}), err
}

func (p *InvocationConfigRepo) toDomain(cfg dao.InvocationConfig) domain.InvocationConfig {
	return domain.InvocationConfig{
		ID:   cfg.ID,
		Name: cfg.Name,
		Biz: domain.BizConfig{
			ID: cfg.ID,
		},
		Description: cfg.Description,
		Ctime:       time.UnixMilli(cfg.Ctime),
		Utime:       time.UnixMilli(cfg.Utime),
	}
}

func (p *InvocationConfigRepo) Count(ctx context.Context) (int, error) {
	return p.dao.Count(ctx)
}

func (p *InvocationConfigRepo) GetVersions(ctx context.Context, invID int64, offset int, limit int) ([]domain.InvocationCfgVersion, error) {
	versions, err := p.dao.GetVersions(ctx, invID, offset, limit)
	return slice.Map(versions, func(idx int, src dao.InvocationConfigVersion) domain.InvocationCfgVersion {
		return p.toDomainVersion(src)
	}), err
}

func (p *InvocationConfigRepo) CountVersions(ctx context.Context, invID int64) (int, error) {
	return p.dao.CountVersions(ctx, invID)
}

func (p *InvocationConfigRepo) SaveVersion(ctx context.Context, version domain.InvocationCfgVersion) (int64, error) {
	return p.dao.SaveVersion(ctx, p.toDAOVersion(version))
}

func (p *InvocationConfigRepo) ActivateVersion(ctx context.Context, id int64) error {
	return p.dao.ActivateVersion(ctx, id)
}
