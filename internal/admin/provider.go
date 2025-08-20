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
	provider := server.Group("/providers")
	provider.POST("/save", ginx.BS(h.SaveProvider))
	provider.POST("/list", ginx.BS(h.ListProviders))
	provider.POST("/detail", ginx.BS(h.ProviderDetail))

	model := server.Group("/models")
	model.POST("/save", ginx.BS(h.SaveModel))
	model.POST("/detail", ginx.BS(h.ModelDetail))
}

func (h *ProviderHandler) SaveProvider(ctx *ginx.Context, req ProviderVO, _ session.Session) (ginx.Result, error) {
	id, err := h.service.SaveProvider(ctx.Request.Context(), newProvider(req))
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{Msg: "OK", Data: id}, nil
}

func (h *ProviderHandler) ListProviders(ctx *ginx.Context, req ListReq, _ session.Session) (ginx.Result, error) {
	provider, total, err := h.service.ListProviders(ctx.Request.Context(), req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: ginx.DataList[ProviderVO]{
			List: slice.Map(provider, func(_ int, src domain.Provider) ProviderVO {
				return h.toProviderVO(src)
			}),
			Total: int(total),
		},
	}, nil
}

func (h *ProviderHandler) toProviderVO(src domain.Provider) ProviderVO {
	return ProviderVO{
		ID:     src.ID,
		Name:   src.Name,
		APIKey: src.APIKey,
		Models: slice.Map(src.Models, func(_ int, src domain.Model) ModelVO {
			return h.toModelVO(src)
		}),
		Ctime: src.Ctime,
		Utime: src.Utime,
	}
}

func (h *ProviderHandler) toModelVO(src domain.Model) ModelVO {
	return ModelVO{
		ID:          src.ID,
		Provider:    h.toProviderVO(src.Provider),
		Name:        src.Name,
		InputPrice:  src.InputPrice,
		OutputPrice: src.OutputPrice,
		PriceMode:   src.PriceMode,
		Ctime:       src.Ctime,
		Utime:       src.Utime,
	}
}

func (h *ProviderHandler) ProviderDetail(ctx *ginx.Context, req IDReq, _ session.Session) (ginx.Result, error) {
	provider, err := h.service.ProviderDetail(ctx.Request.Context(), req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: h.toProviderVO(provider),
	}, nil
}

func (h *ProviderHandler) SaveModel(ctx *ginx.Context, req ModelVO, _ session.Session) (ginx.Result, error) {
	id, err := h.service.SaveModel(ctx.Request.Context(), newModel(req))
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{Msg: "OK", Data: id}, nil
}

func (h *ProviderHandler) ModelDetail(ctx *ginx.Context, req IDReq, _ session.Session) (ginx.Result, error) {
	model, err := h.service.ModelDetail(ctx.Request.Context(), req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: h.toModelVO(model),
	}, nil
}
