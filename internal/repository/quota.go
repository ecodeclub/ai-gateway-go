package repository

import (
	"context"

	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
)

type QuotaRepo struct {
	quota *dao.Quota
}

func NewQuotaRepo(quota *dao.Quota) *QuotaRepo {
	return &QuotaRepo{quota: quota}
}

func (q *QuotaRepo) Get(ctx context.Context) error {
	return nil
}

func (q *QuotaRepo) Deduct(ctx context.Context) error {
}
