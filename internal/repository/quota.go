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

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type QuotaRepo struct {
	dao *dao.QuotaDao
}

func NewQuotaRepo(dao *dao.QuotaDao) *QuotaRepo {
	return &QuotaRepo{dao: dao}
}

func (q *QuotaRepo) AddQuota(ctx context.Context, quota domain.Quota) error {
	return q.dao.AddQuota(ctx, dao.Quota{UID: quota.Uid, Amount: quota.Amount, Key: quota.Key})
}

func (q *QuotaRepo) CreateTempQuota(ctx context.Context, quota domain.TempQuota) error {
	return q.dao.CreateTempQuota(ctx, dao.TempQuota{Amount: quota.Amount, StartTime: quota.StartTime, EndTime: quota.EndTime, Key: quota.Key, UID: quota.Uid})
}

func (q *QuotaRepo) GetQuota(ctx context.Context, uid int64) (domain.Quota, error) {
	quota, err := q.dao.GetQuotaByUid(ctx, uid)
	if err != nil {
		return domain.Quota{}, err
	}
	return domain.Quota{Amount: quota.Amount, Uid: uid}, nil
}

func (q *QuotaRepo) GetTempQuota(ctx context.Context, uid int64) ([]domain.TempQuota, error) {
	tempQuotaList, err := q.dao.GetTempQuotaByUidAndTime(ctx, uid)
	if err != nil {
		return nil, err
	}
	return q.toDomainTempQuota(tempQuotaList), nil
}

func (q *QuotaRepo) Deduct(ctx context.Context, uid int64, amount int64, key string) error {
	return q.dao.Deduct(ctx, uid, amount, key)
}

func (q *QuotaRepo) toDomainTempQuota(tmpQuotaList []dao.TempQuota) []domain.TempQuota {
	return slice.Map[dao.TempQuota, domain.TempQuota](tmpQuotaList, func(idx int, src dao.TempQuota) domain.TempQuota {
		return domain.TempQuota{
			Amount:    src.Amount,
			StartTime: src.StartTime,
			EndTime:   src.EndTime,
		}
	})
}
