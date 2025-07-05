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
	"fmt"

	"github.com/ecodeclub/ai-gateway-go/errs"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

type QuotaService struct {
	repo *repository.QuotaRepo
}

func NewQuotaService(repo *repository.QuotaRepo) *QuotaService {
	return &QuotaService{repo: repo}
}

func (q *QuotaService) SaveQuota(ctx context.Context, quota domain.Quota) error {
	return q.repo.SaveQuota(ctx, quota)
}

func (q *QuotaService) SaveTempQuota(ctx context.Context, quota domain.TempQuota) error {
	return q.repo.SaveTempQuota(ctx, quota)
}

func (q *QuotaService) GetTempQuota(ctx context.Context, uid int64) ([]domain.TempQuota, error) {
	return q.repo.GetTempQuota(ctx, uid)
}

func (q *QuotaService) GetQuota(ctx context.Context, uid int64) (domain.Quota, error) {
	return q.repo.GetQuota(ctx, uid)
}

func (q *QuotaService) Deduct(ctx context.Context, uid int64, amount int64, key string) error {
	key1 := fmt.Sprintf("temp_%s", key)
	key2 := fmt.Sprintf("quota_%s", key)
	tempQuotaList, err := q.repo.GetTempQuota(ctx, uid)
	if err != nil {
		return err
	}
	// 1. 优先扣减临时表
	for i := range tempQuotaList {
		tq := tempQuotaList[i]
		if amount <= 0 {
			break
		}
		deduct := int64(0)
		if tq.Amount >= amount {
			deduct = amount
			amount = 0
		} else {
			deduct = tq.Amount
			amount -= deduct
		}
		tq.Amount -= deduct
		err = q.SaveTempQuota(ctx, domain.TempQuota{
			Uid:    uid,
			Amount: tq.Amount,
			Key:    key1,
		})

		if err != nil {
			return err
		}
	}
	// 扣减完成了
	if amount <= 0 {
		return nil
	}

	quota, err := q.GetQuota(ctx, uid)
	if err != nil {
		return err
	}
	if quota.Amount < amount {
		return errs.ErrNoAmount
	}

	return q.SaveQuota(ctx, domain.Quota{
		Uid:    uid,
		Amount: amount,
		Key:    key2,
	})
}
