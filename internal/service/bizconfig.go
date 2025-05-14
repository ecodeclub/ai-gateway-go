package service

import (
	"context"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

type BizConfigService interface {
	Create(ctx context.Context, config domain.BizConfig) (domain.BizConfig, error)
	GetByID(ctx context.Context, id int64) (domain.BizConfig, error)
	Update(ctx context.Context, config domain.BizConfig) error
	Delete(ctx context.Context, id string) error
}

type bizConfigService struct {
	repo *repository.BizConfigRepository
}

func NewBizConfigService(repo *repository.BizConfigRepository) BizConfigService {
	return &bizConfigService{
		repo: repo,
	}
}

func (s *bizConfigService) Create(ctx context.Context, req domain.BizConfig) (domain.BizConfig, error) {
	config := domain.BizConfig{
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Config:    req.Config,
	}

	created, err := s.repo.Create(ctx, config)
	if err != nil {
		return domain.BizConfig{}, err
	}

	return created, nil
}

func (s *bizConfigService) GetByID(ctx context.Context, id int64) (domain.BizConfig, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *bizConfigService) Update(ctx context.Context, config domain.BizConfig) error {
	return s.repo.Update(ctx, config)
}

func (s *bizConfigService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
