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

type ProviderRepository struct {
	dao *dao.ProviderDAO
}

func NewProviderRepository(dao *dao.ProviderDAO) *ProviderRepository {
	return &ProviderRepository{
		dao: dao,
	}
}

func (r *ProviderRepository) SaveProvider(ctx context.Context, provider domain.Provider) (int64, error) {
	return r.dao.SaveProvider(ctx, dao.Provider{
		ID:     provider.ID,
		Name:   provider.Name,
		APIKey: provider.APIKey,
	})
}

func (r *ProviderRepository) ListProviders(ctx context.Context, offset, limit int) ([]domain.Provider, error) {
	providers, err := r.dao.ListProviders(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map(providers, func(_ int, src dao.Provider) domain.Provider {
		return r.toDomainProvider(src, nil)
	}), nil
}

func (r *ProviderRepository) toDomainProvider(provider dao.Provider, models []dao.Model) domain.Provider {
	return domain.Provider{
		ID:     provider.ID,
		Name:   provider.Name,
		APIKey: provider.APIKey,
		Models: slice.Map(models, func(_ int, src dao.Model) domain.Model {
			return r.toDomainModel(src, provider)
		}),
		Ctime: provider.Ctime,
		Utime: provider.Utime,
	}
}

func (r *ProviderRepository) toDomainModel(m dao.Model, p dao.Provider) domain.Model {
	return domain.Model{
		ID:          m.ID,
		Name:        m.Name,
		Provider:    r.toDomainProvider(p, nil),
		OutputPrice: m.OutputPrice,
		InputPrice:  m.InputPrice,
		PriceMode:   m.PriceMode,
		Ctime:       m.Ctime,
		Utime:       m.Utime,
	}
}

func (r *ProviderRepository) CountProviders(ctx context.Context) (int64, error) {
	return r.dao.CountProviders(ctx)
}

func (r *ProviderRepository) GetProvider(ctx context.Context, id int64) (domain.Provider, error) {
	provider, err := r.dao.GetProvider(ctx, id)
	if err != nil {
		return domain.Provider{}, err
	}
	models, err := r.dao.GetModelsByPid(ctx, provider.ID)
	if err != nil {
		return domain.Provider{}, err
	}
	return r.toDomainProvider(provider, models), nil
}

func (r *ProviderRepository) SaveModel(ctx context.Context, model domain.Model) (int64, error) {
	return r.dao.SaveModel(ctx, dao.Model{
		ID:          model.ID,
		Name:        model.Name,
		Pid:         model.Provider.ID,
		InputPrice:  model.InputPrice,
		OutputPrice: model.OutputPrice,
		PriceMode:   model.PriceMode,
	})
}

func (r *ProviderRepository) GetModel(ctx context.Context, id int64) (domain.Model, error) {
	model, err := r.dao.GetModel(ctx, id)
	if err != nil {
		return domain.Model{}, err
	}
	provider, err := r.dao.GetProvider(ctx, model.Pid)
	if err != nil {
		return domain.Model{}, err
	}
	return r.toDomainModel(model, provider), nil
}
