package repository

import (
	"context"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/cache"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type ProviderRepo struct {
	dao   *dao.ProviderDao
	cache *cache.ProviderCache
}

func NewProviderRepo(dao *dao.ProviderDao, cache *cache.ProviderCache) *ProviderRepo {
	return &ProviderRepo{dao: dao, cache: cache}
}

func (p *ProviderRepo) SaveProvider(ctx context.Context, provider domain.Provider) (int64, error) {
	id, err := p.dao.SaveProvider(ctx, dao.Provider{
		Id:     provider.Id,
		Name:   provider.Name,
		APIKey: provider.ApiKey,
	})

	if err != nil {
		return 0, err
	}
	err = p.cache.AddProvider(ctx, cache.Provider{
		Id:     id,
		Name:   provider.Name,
		APIKey: provider.ApiKey,
	})
	return id, err
}

func (p *ProviderRepo) SaveModel(ctx context.Context, model domain.Model) (int64, error) {
	id, err := p.dao.SaveModel(ctx, dao.Model{
		Id:          model.Id,
		Name:        model.Name,
		InputPrice:  model.InputPrice,
		OutputPrice: model.OutPutPrice,
		PriceMode:   model.PriceMode,
	})
	if err != nil {
		return 0, err
	}

	err = p.cache.AddModel(ctx, cache.Model{
		Id:          id,
		Pid:         model.Pid,
		Name:        model.Name,
		InputPrice:  model.InputPrice,
		OutputPrice: model.OutPutPrice,
		PriceMode:   model.PriceMode,
	})

	return id, err
}

func (p *ProviderRepo) GetProviders(ctx context.Context) ([]domain.Provider, error) {
	var providers []domain.Provider

	cacheProvider, err := p.cache.GetAllProvider(ctx)
	if err != nil {
		providers, err = p.getProvider(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		providers = p.toProvider(cacheProvider)
	}

	var models []domain.Model
	for _, provider := range providers {
		cacheModels, err := p.cache.GetModelListByPid(ctx, provider.Id)
		if err != nil {
			models, err = p.getModel(ctx, provider.Id)
			if err != nil {
				return nil, err
			}
		} else {
			provider.Models = p.toModel(cacheModels)
		}
		provider.Models = models
	}

	return providers, nil
}

func (p *ProviderRepo) getProvider(ctx context.Context) ([]domain.Provider, error) {
	providers, err := p.dao.GetAllProviders(ctx)
	if err != nil {
		return nil, err
	}
	return p.toDomainProvider(providers), nil
}

func (p *ProviderRepo) getModel(ctx context.Context, pid int64) ([]domain.Model, error) {
	models, err := p.dao.GetModels(ctx, pid)
	if err != nil {
		return nil, err
	}
	return p.toDomainModel(models), err
}

func (p *ProviderRepo) toProvider(list []cache.Provider) []domain.Provider {
	return slice.Map[cache.Provider, domain.Provider](list, func(idx int, src cache.Provider) domain.Provider {
		return domain.Provider{
			Id:     src.Id,
			Name:   src.Name,
			ApiKey: src.APIKey,
		}
	})
}

func (p *ProviderRepo) toDomainProvider(list []dao.Provider) []domain.Provider {
	return slice.Map[dao.Provider, domain.Provider](list, func(idx int, src dao.Provider) domain.Provider {
		return domain.Provider{
			Id:     src.Id,
			Name:   src.Name,
			ApiKey: src.APIKey,
		}
	})
}

func (p *ProviderRepo) toDomainModel(list []dao.Model) []domain.Model {
	return slice.Map[dao.Model, domain.Model](list, func(idx int, src dao.Model) domain.Model {
		return domain.Model{
			Id:          src.Id,
			Pid:         src.Pid,
			OutPutPrice: src.OutputPrice,
			InputPrice:  src.InputPrice,
			PriceMode:   src.PriceMode,
		}
	})
}

func (p *ProviderRepo) toModel(list []cache.Model) []domain.Model {
	return slice.Map[cache.Model, domain.Model](list, func(idx int, src cache.Model) domain.Model {
		return domain.Model{
			Id:          src.Id,
			Pid:         src.Pid,
			InputPrice:  src.InputPrice,
			OutPutPrice: src.OutputPrice,
			PriceMode:   src.PriceMode,
		}
	})
}
