package service

import (
	"context"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

type ProviderService struct {
	repo *repository.ProviderRepo
}

func NewProviderService(repo *repository.ProviderRepo) *ProviderService {
	return &ProviderService{repo: repo}
}

func (p *ProviderService) SaveProvider(ctx context.Context, provider domain.Provider) (int64, error) {
	return p.repo.SaveProvider(ctx, provider)
}

func (p *ProviderService) SaveModel(ctx context.Context, model domain.Model) (int64, error) {
	return p.repo.SaveModel(ctx, model)
}

func (p *ProviderService) GetProviders(ctx context.Context) ([]domain.Provider, error) {
	return p.repo.GetProviders(ctx)
}
