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
		Id:     provider.ID,
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
		Id:          model.ID,
		Name:        model.Name,
		InputPrice:  model.InputPrice,
		OutputPrice: model.OutputPrice,
		PriceMode:   model.PriceMode,
	})
	if err != nil {
		return 0, err
	}

	err = p.cache.AddModel(ctx, cache.Model{
		Id:          id,
		Pid:         model.Provider.ID,
		Name:        model.Name,
		InputPrice:  model.InputPrice,
		OutputPrice: model.OutputPrice,
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
		cacheModels, err := p.cache.GetModelListByPid(ctx, provider.ID)
		if err != nil {
			models, err = p.getModel(ctx, provider.ID)
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
			ID:     src.Id,
			Name:   src.Name,
			ApiKey: src.APIKey,
		}
	})
}

func (p *ProviderRepo) toDomainProvider(list []dao.Provider) []domain.Provider {
	return slice.Map[dao.Provider, domain.Provider](list, func(idx int, src dao.Provider) domain.Provider {
		return domain.Provider{
			ID:     src.Id,
			Name:   src.Name,
			ApiKey: src.APIKey,
		}
	})
}

func (p *ProviderRepo) toDomainModel(list []dao.Model) []domain.Model {
	return slice.Map[dao.Model, domain.Model](list, func(idx int, src dao.Model) domain.Model {
		return domain.Model{
			ID:          src.Id,
			Provider:    domain.Provider{ID: src.Pid},
			OutputPrice: src.OutputPrice,
			InputPrice:  src.InputPrice,
			PriceMode:   src.PriceMode,
		}
	})
}

func (p *ProviderRepo) toModel(list []cache.Model) []domain.Model {
	return slice.Map[cache.Model, domain.Model](list, func(idx int, src cache.Model) domain.Model {
		return domain.Model{
			ID:          src.Id,
			Provider:    domain.Provider{ID: src.Pid},
			InputPrice:  src.InputPrice,
			OutputPrice: src.OutputPrice,
			PriceMode:   src.PriceMode,
		}
	})
}
