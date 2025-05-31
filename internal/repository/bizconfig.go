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
	"errors"
	"time"

	"github.com/ecodeclub/ai-gateway-go/errs"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
)

type BizConfigRepository struct {
	dao *dao.BizConfigDAO
}

func NewBizConfigRepository(dao *dao.BizConfigDAO) *BizConfigRepository {
	return &BizConfigRepository{dao: dao}
}

func (r *BizConfigRepository) Create(ctx context.Context, config domain.BizConfig) (domain.BizConfig, error) {
	daoBC, err := r.dao.Insert(ctx, toDAOConfig(config))
	if err != nil {
		return domain.BizConfig{}, err
	}
	return fromDAOConfig(daoBC), nil
}

func (r *BizConfigRepository) GetByID(ctx context.Context, id int64) (domain.BizConfig, error) {
	bc, err := r.dao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrBizConfigNotFound) {
			return domain.BizConfig{}, errs.ErrBizConfigNotFound
		}
		return domain.BizConfig{}, err
	}
	return fromDAOConfig(bc), nil
}

func (r *BizConfigRepository) Update(ctx context.Context, config domain.BizConfig) error {
	return r.dao.Update(ctx, toDAOConfig(config))
}

func (r *BizConfigRepository) Delete(ctx context.Context, id string) error {
	return r.dao.Delete(ctx, id)
}

func toDAOConfig(config domain.BizConfig) *dao.BizConfig {
	return &dao.BizConfig{
		ID:        config.ID,
		OwnerID:   config.OwnerID,
		OwnerType: config.OwnerType,
		Config:    config.Config,
		Ctime:     config.Ctime.UnixMilli(),
		Utime:     config.Utime.UnixMilli(),
	}
}

func fromDAOConfig(bc dao.BizConfig) domain.BizConfig {
	return domain.BizConfig{
		ID:        bc.ID,
		OwnerID:   bc.OwnerID,
		OwnerType: bc.OwnerType,
		Config:    bc.Config,
		Ctime:     time.UnixMilli(bc.Ctime),
		Utime:     time.UnixMilli(bc.Utime),
	}
}
