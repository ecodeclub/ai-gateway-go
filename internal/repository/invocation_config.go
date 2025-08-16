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

package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/ecodeclub/ekit/slice"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
)

type InvocationConfigRepo struct {
	dao *dao.InvocationConfigDAO
}

func NewInvocationConfigRepo(dao *dao.InvocationConfigDAO) *InvocationConfigRepo {
	return &InvocationConfigRepo{dao: dao}
}

func (p *InvocationConfigRepo) Save(ctx context.Context, cfg domain.InvocationConfig) (int64, error) {
	return p.dao.Save(ctx, p.toEntity(cfg))
}

func (p *InvocationConfigRepo) toEntity(src domain.InvocationConfig) dao.InvocationConfig {
	return dao.InvocationConfig{
		ID:          src.ID,
		Name:        src.Name,
		BizID:       src.Biz.ID,
		Description: src.Description,
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
			ID: cfg.BizID,
		},
		Description: cfg.Description,
		Ctime:       time.UnixMilli(cfg.Ctime),
		Utime:       time.UnixMilli(cfg.Utime),
	}
}

func (p *InvocationConfigRepo) Count(ctx context.Context) (int, error) {
	return p.dao.Count(ctx)
}

func (p *InvocationConfigRepo) Get(ctx context.Context, id int64) (domain.InvocationConfig, error) {
	cfg, err := p.dao.GetByID(ctx, id)
	if err != nil {
		return domain.InvocationConfig{}, err
	}
	return p.toDomain(cfg), nil
}

func (p *InvocationConfigRepo) SaveVersion(ctx context.Context, version domain.InvocationConfigVersion) (int64, error) {
	return p.dao.SaveVersion(ctx, p.toVersionEntity(version))
}

func (p *InvocationConfigRepo) toVersionEntity(src domain.InvocationConfigVersion) dao.InvocationConfigVersion {
	return dao.InvocationConfigVersion{
		ID:           src.ID,
		InvID:        src.Config.ID,
		ModelID:      src.Model.ID,
		Version:      src.Version,
		Prompt:       src.Prompt,
		SystemPrompt: src.SystemPrompt,
		JSONSchema:   sql.Null[string]{V: src.JSONSchema, Valid: src.JSONSchema != ""},
		Temperature:  src.Temperature,
		TopP:         src.TopP,
		MaxTokens:    src.MaxTokens,
		Status:       src.Status.String(),
	}
}

func (p *InvocationConfigRepo) ListVersions(ctx context.Context, invID int64, offset int, limit int) ([]domain.InvocationConfigVersion, error) {
	versions, err := p.dao.ListVersions(ctx, invID, offset, limit)
	return slice.Map(versions, func(idx int, src dao.InvocationConfigVersion) domain.InvocationConfigVersion {
		return p.toDomainVersion(src)
	}), err
}

func (p *InvocationConfigRepo) toDomainVersion(v dao.InvocationConfigVersion) domain.InvocationConfigVersion {
	var jsonSchema string
	if v.JSONSchema.Valid {
		jsonSchema = v.JSONSchema.V
	}
	return domain.InvocationConfigVersion{
		ID:           v.ID,
		Config:       domain.InvocationConfig{ID: v.InvID},
		Version:      v.Version,
		Model:        domain.Model{ID: v.ModelID},
		Prompt:       v.Prompt,
		SystemPrompt: v.SystemPrompt,
		JSONSchema:   jsonSchema,
		Temperature:  v.Temperature,
		TopP:         v.TopP,
		MaxTokens:    v.MaxTokens,
		Status:       domain.InvocationConfigVersionStatus(v.Status),
		Ctime:        time.UnixMilli(v.Ctime),
		Utime:        time.UnixMilli(v.Utime),
	}
}

func (p *InvocationConfigRepo) CountVersions(ctx context.Context, invID int64) (int, error) {
	return p.dao.CountVersions(ctx, invID)
}

func (p *InvocationConfigRepo) GetVersionByID(ctx context.Context, id int64) (domain.InvocationConfigVersion, error) {
	res, err := p.dao.GetVersionByD(ctx, id)
	if err != nil {
		return domain.InvocationConfigVersion{}, err
	}
	return p.toDomainVersion(res), nil
}

func (p *InvocationConfigRepo) ActivateVersion(ctx context.Context, id int64) error {
	return p.dao.ActivateVersion(ctx, id)
}
