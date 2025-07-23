// Copyright 2023 ecodeclub
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

package admin

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gotomicro/ego/server/egin"
)

type ProviderHandler struct {
	service *service.ProviderService
}

func NewProviderHandler(svc *service.ProviderService) *ProviderHandler {
	return &ProviderHandler{service: svc}
}

func (h *ProviderHandler) PrivateRoutes(server *egin.Component) {
	g := server.Group("/providers")
	g.POST("/all", ginx.S(h.AllProviders))
	g.POST("/model/save", ginx.BS(h.SaveModel))
	g.POST("/save", ginx.BS(h.SaveProvider))
}

func (h *ProviderHandler) AllProviders(ctx *ginx.Context, sess session.Session) (ginx.Result, error) {
	provider, err := h.service.GetProviders(ctx)
	if err != nil {
		return ginx.Result{Code: 500, Msg: "内部错误"}, ginx.ErrNoResponse
	}
	list := h.toProviderList(provider)

	return ginx.Result{Data: list}, nil
}

func (h *ProviderHandler) SaveProvider(ctx *ginx.Context, req ProviderVO, sess session.Session) (ginx.Result, error) {
	id, err := h.service.SaveProvider(ctx, h.toDomainProvider(req))
	if err != nil {
		return ginx.Result{Code: 500, Msg: "内部错误"}, ginx.ErrNoResponse
	}
	return ginx.Result{Code: 200, Data: id}, nil
}

func (h *ProviderHandler) SaveModel(ctx *ginx.Context, req ModelVO, sess session.Session) (ginx.Result, error) {
	id, err := h.service.SaveModel(ctx, h.toDomainModel(req))
	if err != nil {
		return ginx.Result{Code: 500, Msg: "内部错误"}, ginx.ErrNoResponse
	}
	return ginx.Result{Code: 200, Data: id}, nil
}

func (h *ProviderHandler) toProviderList(providers []domain.Provider) []ProviderVO {
	return slice.Map(providers, func(idx int, src domain.Provider) ProviderVO {
		return h.toProviderVO(src)
	})
}

func (h *ProviderHandler) toProviderVO(provider domain.Provider) ProviderVO {
	return ProviderVO{
		ID:     provider.ID,
		Name:   provider.Name,
		Models: h.toModelVO(provider.Models),
	}
}

func (h *ProviderHandler) toModelVO(models []domain.Model) []ModelVO {
	return slice.Map[domain.Model, ModelVO](models, func(idx int, src domain.Model) ModelVO {
		return ModelVO{
			ID: src.ID,
		}
	})
}

func (h *ProviderHandler) toDomainProvider(provider ProviderVO) domain.Provider {
	return domain.Provider{
		ID:     provider.ID,
		Name:   provider.Name,
		ApiKey: provider.ApiKey,
	}
}

func (h *ProviderHandler) toDomainModel(model ModelVO) domain.Model {
	return domain.Model{
		ID:          model.ID,
		Name:        model.Name,
		InputPrice:  model.InputPrice,
		OutputPrice: model.OutputPrice,
		Provider:    domain.Provider{ID: model.Pid},
	}
}
