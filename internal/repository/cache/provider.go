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

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	ProviderAllKey = "provider:all"
	ProviderKey    = "provider:%d"
)

type ProviderCache struct {
	rdb        redis.Cmdable
	expiration time.Duration
}

func NewProviderCache(rdb redis.Cmdable) *ProviderCache {
	// 默认永不过期
	return &ProviderCache{rdb: rdb, expiration: 0}
}

func (p *ProviderCache) SetProvider(ctx context.Context, provider Provider) error {
	key := p.providerKey(provider.ID)
	jsonData, err := json.Marshal(provider)
	if err != nil {
		return err
	}
	return p.rdb.Set(ctx, key, jsonData, p.expiration).Err()
}

func (p *ProviderCache) GetProvider(ctx context.Context, id int64) (Provider, error) {
	bs, err := p.rdb.Get(ctx, p.providerKey(id)).Bytes()
	if err != nil {
		return Provider{}, err
	}
	var provider Provider
	err = json.Unmarshal(bs, &provider)
	return provider, err
}

func (p *ProviderCache) providerKey(id int64) string {
	return fmt.Sprintf("provider:%d", id)
}

func (p *ProviderCache) AddModel(ctx context.Context, model Model) error {
	field := p.getModelField(model.ID)
	jsonData, err := json.Marshal(model)
	if err != nil {
		return err
	}
	key := fmt.Sprintf(ProviderKey, model.Pid)
	return p.rdb.HSet(ctx, key, field, jsonData).Err()
}

func (p *ProviderCache) getModelField(id int64) string {
	return fmt.Sprintf("model:%d", id)
}

func (p *ProviderCache) Reload(ctx context.Context, providers []Provider, models []Model) error {
	pipe := p.rdb.TxPipeline()

	pipe.Del(ctx, ProviderAllKey)

	providerFields := make(map[string]interface{})
	for _, provider := range providers {
		data, err := json.Marshal(provider)
		if err != nil {
			return err
		}
		field := p.providerKey(provider.ID)
		providerFields[field] = data
		pipe.Del(ctx, fmt.Sprintf(ProviderKey, provider.ID))
	}
	if len(providerFields) > 0 {
		pipe.HSet(ctx, ProviderAllKey, providerFields)
	}

	for _, model := range models {
		data, err := json.Marshal(model)
		if err != nil {
			return err
		}
		field := p.getModelField(model.ID)
		pipe.HSet(ctx, fmt.Sprintf(ProviderKey, model.Pid), field, data)
	}

	_, err := pipe.Exec(ctx)
	return err
}

type Provider struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	APIKey string `json:"apiKey"`
}

type Model struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Pid         int64  `json:"pid,omitempty"`
	InputPrice  int64  `json:"inputPrice,omitempty"`
	OutputPrice int64  `json:"outputPrice,omitempty"`
	PriceMode   string `json:"priceMode,omitempty"`
}
