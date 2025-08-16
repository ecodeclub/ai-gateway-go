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
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type BizConfigRepository struct {
	dao *dao.BizConfigDAO
}

func NewBizConfigRepository(dao *dao.BizConfigDAO) *BizConfigRepository {
	return &BizConfigRepository{dao: dao}
}

func (r *BizConfigRepository) Save(ctx context.Context, config domain.BizConfig) (int64, error) {
	return r.dao.Save(ctx, toEntity(config))
}

func toEntity(b domain.BizConfig) dao.BizConfig {
	return dao.BizConfig{
		ID:        b.ID,
		Name:      b.Name,
		OwnerID:   b.OwnerID,
		OwnerType: b.OwnerType,
		Config:    b.Config,
		Ctime:     b.Ctime.UnixMilli(),
		Utime:     b.Utime.UnixMilli(),
	}
}

func (r *BizConfigRepository) List(ctx context.Context, offset, limit int) ([]domain.BizConfig, error) {
	configs, err := r.dao.List(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map(configs, func(idx int, src dao.BizConfig) domain.BizConfig {
		return r.toDomain(src)
	}), err
}

func (r *BizConfigRepository) toDomain(bc dao.BizConfig) domain.BizConfig {
	return domain.BizConfig{
		ID:        bc.ID,
		Name:      bc.Name,
		OwnerID:   bc.OwnerID,
		OwnerType: bc.OwnerType,
		Config:    bc.Config,
		Ctime:     time.UnixMilli(bc.Ctime),
		Utime:     time.UnixMilli(bc.Utime),
	}
}

func (r *BizConfigRepository) Count(ctx context.Context) (int64, error) {
	return r.dao.Count(ctx)
}

func (r *BizConfigRepository) GetByID(ctx context.Context, id int64) (domain.BizConfig, error) {
	bc, err := r.dao.GetByID(ctx, id)
	if err != nil {
		return domain.BizConfig{}, err
	}
	return r.toDomain(bc), nil
}
