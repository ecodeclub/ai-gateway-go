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
	"github.com/ecodeclub/ekit/slice"
)

type ProviderVO struct {
	ID     int64     `json:"id"`
	Name   string    `json:"name"`
	APIKey string    `json:"apiKey,omitempty"`
	Models []ModelVO `json:"models,omitempty"`
	Ctime  int64     `json:"ctime"`
	Utime  int64     `json:"utime"`
}

func newProvider(req ProviderVO) domain.Provider {
	return domain.Provider{
		ID:     req.ID,
		Name:   req.Name,
		APIKey: req.APIKey,
		Models: slice.Map(req.Models, func(_ int, src ModelVO) domain.Model {
			return newModel(src)
		}),
		Ctime: req.Ctime,
		Utime: req.Utime,
	}
}

type ModelVO struct {
	ID          int64      `json:"id"`
	Provider    ProviderVO `json:"provider"`
	Name        string     `json:"name"`
	InputPrice  int64      `json:"inputPrice"`
	OutputPrice int64      `json:"outputPrice"`
	PriceMode   string     `json:"priceMode"`
	Ctime       int64      `json:"ctime"`
	Utime       int64      `json:"utime"`
}

func newModel(req ModelVO) domain.Model {
	return domain.Model{
		ID: req.ID,
		Provider: domain.Provider{
			ID:     req.Provider.ID,
			Name:   req.Provider.Name,
			APIKey: req.Provider.APIKey,
			Ctime:  req.Provider.Ctime,
			Utime:  req.Provider.Utime,
		},
		Name:        req.Name,
		InputPrice:  req.InputPrice,
		OutputPrice: req.OutputPrice,
		PriceMode:   req.PriceMode,
		Ctime:       req.Ctime,
		Utime:       req.Utime,
	}
}
