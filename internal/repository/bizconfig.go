package repository

import (
	"context"
	"errors"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"gorm.io/gorm"
)

var ErrBizConfigNotFound = dao.ErrBizConfigNotFound
var ErrQuotaExhausted = dao.ErrQuotaExhausted

type BizConfigRepository struct {
	dao *dao.BizConfigDAO
}

func NewBizConfigRepository(dao *dao.BizConfigDAO) *BizConfigRepository {
	return &BizConfigRepository{dao: dao}
}

func (r *BizConfigRepository) Create(ctx context.Context, config domain.BizConfig) (domain.BizConfig, error) {
	daoBC, err := r.dao.Insert(ctx, &dao.BizConfig{
		ID:        config.ID,
		OwnerID:   config.OwnerID,
		OwnerType: config.OwnerType,
		Token:     config.Token,
		Config:    config.Config,
		Quota:     config.Quota,
		UsedQuota: config.UsedQuota,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	})
	if err != nil {
		return domain.BizConfig{}, err
	}

	return domain.BizConfig{
		ID:        daoBC.ID,
		OwnerID:   daoBC.OwnerID,
		OwnerType: daoBC.OwnerType,
		Token:     daoBC.Token,
		Config:    daoBC.Config,
		Quota:     daoBC.Quota,
		UsedQuota: daoBC.UsedQuota,
		CreatedAt: daoBC.CreatedAt,
		UpdatedAt: daoBC.UpdatedAt,
	}, nil
}

func (r *BizConfigRepository) GetByID(ctx context.Context, id string) (domain.BizConfig, error) {
	bc, err := r.dao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.BizConfig{}, ErrBizConfigNotFound
		}
		return domain.BizConfig{}, err
	}

	return domain.BizConfig{
		ID:        bc.ID,
		OwnerID:   bc.OwnerID,
		OwnerType: bc.OwnerType,
		Token:     bc.Token,
		Config:    bc.Config,
		Quota:     bc.Quota,
		UsedQuota: bc.UsedQuota,
		CreatedAt: bc.CreatedAt,
		UpdatedAt: bc.UpdatedAt,
	}, nil
}

func (r *BizConfigRepository) Update(ctx context.Context, config domain.BizConfig) error {
	return r.dao.Update(ctx, &dao.BizConfig{
		ID:        config.ID,
		OwnerID:   config.OwnerID,
		OwnerType: config.OwnerType,
		Token:     config.Token,
		Config:    config.Config,
		Quota:     config.Quota,
		UsedQuota: config.UsedQuota,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	})
}

func (r *BizConfigRepository) Delete(ctx context.Context, id string) error {
	return r.dao.Delete(ctx, id)
}

func (r *BizConfigRepository) List(ctx context.Context, ownerID int64, ownerType string, page, pageSize int) ([]domain.BizConfig, int, error) {
	return r.dao.List(ctx, ownerID, ownerType, page, pageSize)
}

func (r *BizConfigRepository) CheckAndUpdateQuota(ctx context.Context, id string, requiredQuota int64) (bool, int64, error) {
	return r.dao.CheckAndUpdateQuota(ctx, id, requiredQuota)
}

func (r *BizConfigRepository) GetRemainingQuota(ctx context.Context, id string) (int64, error) {
	return r.dao.GetRemainingQuota(ctx, id)
}
