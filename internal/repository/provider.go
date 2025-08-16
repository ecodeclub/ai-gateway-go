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

	"github.com/ecodeclub/ekit/mapx"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/cache"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"golang.org/x/sync/errgroup"
)

type ProviderRepo struct {
	dao    *dao.ProviderDao
	cache  *cache.ProviderCache
	logger *elog.Component
}

func NewProviderRepo(dao *dao.ProviderDao, cache *cache.ProviderCache) *ProviderRepo {
	provider := &ProviderRepo{dao: dao, cache: cache, logger: elog.DefaultLogger.With(elog.FieldComponent("ProviderRepo"))}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := provider.ReloadCache(ctx); err != nil {
			provider.logger.Error("异步预热缓存失败", elog.FieldErr(err))
		}
	}()
	return provider
}

func (p *ProviderRepo) SaveProvider(ctx context.Context, provider domain.Provider) (int64, error) {
	id, err := p.dao.SaveProvider(ctx, dao.Provider{
		ID:     provider.ID,
		Name:   provider.Name,
		APIKey: provider.ApiKey,
	})

	if err != nil {
		return 0, err
	}
	err = p.cache.SetProvider(ctx, cache.Provider{
		ID:     id,
		Name:   provider.Name,
		APIKey: provider.ApiKey,
	})
	if err != nil {
		p.logger.Error("更新 redis 失败", elog.Any("err", err), elog.Int("id", int(provider.ID)))
	}
	return id, nil
}

func (p *ProviderRepo) SaveModel(ctx context.Context, model domain.Model) (int64, error) {
	id, err := p.dao.SaveModel(ctx, dao.Model{
		ID:          model.ID,
		Name:        model.Name,
		Pid:         model.Provider.ID,
		InputPrice:  model.InputPrice,
		OutputPrice: model.OutputPrice,
		PriceMode:   model.PriceMode,
	})
	if err != nil {
		return 0, err
	}

	err = p.cache.AddModel(ctx, cache.Model{
		ID:          id,
		Pid:         model.Provider.ID,
		Name:        model.Name,
		InputPrice:  model.InputPrice,
		OutputPrice: model.OutputPrice,
		PriceMode:   model.PriceMode,
	})
	if err != nil {
		p.logger.Error("更新 redis 失败", elog.Any("err", err), elog.Int("id", int(model.ID)))
	}
	return id, nil
}

func (p *ProviderRepo) GetProvider(ctx context.Context, id int64) (domain.Provider, error) {
	provider, err := p.dao.GetProvider(ctx, id)
	if err != nil {
		return domain.Provider{}, err
	}
	return domain.Provider{ID: provider.ID, Name: provider.Name, ApiKey: provider.APIKey}, nil
}

func (p *ProviderRepo) GetModel(ctx context.Context, id int64) (domain.Model, error) {
	model, err := p.dao.GetModel(ctx, id)
	if err != nil {
		return domain.Model{}, err
	}
	return domain.Model{ID: model.ID, Name: model.Name, InputPrice: model.InputPrice, OutputPrice: model.OutputPrice}, nil
}

// GetAll 不会经过缓存，因为它目前只会用于管理后台。
func (p *ProviderRepo) GetAll(ctx context.Context) ([]domain.Provider, error) {
	var (
		eg        errgroup.Group
		providers []domain.Provider
		models    []domain.Model
	)
	eg.Go(func() error {
		var err error
		providers, err = p.getAllProviders(ctx)
		return err
	})

	eg.Go(func() error {
		var err error
		models, err = p.getAllModels(ctx)
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	res := mapx.NewMultiBuiltinMap[int64, domain.Model](len(models))
	for _, model := range models {
		_ = res.Put(model.Provider.ID, model)
	}

	for _, provider := range providers {
		ms, _ := res.Get(provider.ID)
		provider.Models = ms
	}
	return providers, nil
}

func (p *ProviderRepo) getAllProviders(ctx context.Context) ([]domain.Provider, error) {
	providers, err := p.dao.GetAllProviders(ctx)
	if err != nil {
		return nil, err
	}
	return p.toDomainProvider(providers), nil
}

func (p *ProviderRepo) getAllModels(ctx context.Context) ([]domain.Model, error) {
	models, err := p.dao.GetAllModel(ctx)
	if err != nil {
		return nil, err
	}
	return p.toDomainModels(models), err
}

func (p *ProviderRepo) ReloadCache(ctx context.Context) error {
	var (
		eg        errgroup.Group
		providers []domain.Provider
		models    []domain.Model
	)

	eg.Go(func() error {
		var err error
		providers, err = p.getAllProviders(ctx)
		return err
	})

	eg.Go(func() error {
		var err error
		models, err = p.getAllModels(ctx)
		return err
	})

	if err := eg.Wait(); err != nil {
		return err
	}
	return p.cache.Reload(ctx, p.daoToProvider(providers), p.daoToModels(models))
}

func (p *ProviderRepo) toDomainProvider(list []dao.Provider) []domain.Provider {
	return slice.Map[dao.Provider, domain.Provider](list, func(idx int, src dao.Provider) domain.Provider {
		return domain.Provider{
			ID:     src.ID,
			Name:   src.Name,
			ApiKey: src.APIKey,
		}
	})
}

func (p *ProviderRepo) toDomainModels(list []dao.Model) []domain.Model {
	return slice.Map[dao.Model, domain.Model](list, func(idx int, src dao.Model) domain.Model {
		return p.toDomainModel(src)
	})
}

func (p *ProviderRepo) toDomainModel(m dao.Model) domain.Model {
	return domain.Model{
		ID:          m.ID,
		Name:        m.Name,
		Provider:    domain.Provider{ID: m.Pid},
		OutputPrice: m.OutputPrice,
		InputPrice:  m.InputPrice,
		PriceMode:   m.PriceMode,
	}
}

func (p *ProviderRepo) daoToProvider(providers []domain.Provider) []cache.Provider {
	return slice.Map[domain.Provider, cache.Provider](providers, func(idx int, src domain.Provider) cache.Provider {
		return cache.Provider{
			ID:     src.ID,
			Name:   src.Name,
			APIKey: src.ApiKey,
		}
	})
}

func (p *ProviderRepo) daoToModels(models []domain.Model) []cache.Model {
	return slice.Map[domain.Model, cache.Model](models, func(idx int, src domain.Model) cache.Model {
		return cache.Model{
			ID:          src.ID,
			Name:        src.Name,
			Pid:         src.Provider.ID,
			InputPrice:  src.InputPrice,
			OutputPrice: src.OutputPrice,
			PriceMode:   src.PriceMode,
		}
	})
}
