package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	ProviderAllKey = "provider:all"
	ProviderKey    = "provider:%d"
)

type ProviderCache struct {
	rdb redis.Cmdable
}

func NewProviderCache(rdb redis.Cmdable) *ProviderCache {
	return &ProviderCache{rdb: rdb}
}

func (p *ProviderCache) AddProvider(ctx context.Context, provider Provider) error {
	field := p.getProviderField(provider)
	jsonData, err := json.Marshal(provider)
	if err != nil {
		return err
	}
	return p.rdb.HSet(ctx, ProviderAllKey, field, jsonData).Err()
}
func (p *ProviderCache) getProviderField(provider Provider) string {
	return fmt.Sprintf("%s:%d", provider.Name, provider.Id)
}

func (p *ProviderCache) AddModel(ctx context.Context, model Model) error {
	field := p.modelKey(model.Name, model.Id)
	jsonData, err := json.Marshal(model)
	if err != nil {
		return err
	}
	key := fmt.Sprintf(ProviderKey, model.Pid)
	return p.rdb.HSet(ctx, key, field, jsonData).Err()
}

func (p *ProviderCache) modelKey(name string, id int64) string {
	return fmt.Sprintf("%s:%d", name, id)
}

func (p *ProviderCache) GetAllProvider(ctx context.Context) ([]Provider, error) {
	result, err := p.rdb.HGetAll(ctx, ProviderAllKey).Result()
	if err != nil {
		return nil, err
	}
	providers := make([]Provider, 0, len(result))
	for _, val := range result {
		var provider Provider
		if err := json.Unmarshal([]byte(val), &provider); err != nil {
			continue
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

func (p *ProviderCache) GetModelListByPid(ctx context.Context, pid int64) ([]Model, error) {
	result, err := p.rdb.HGetAll(ctx, fmt.Sprintf(ProviderKey, pid)).Result()
	if err != nil {
		return nil, err
	}

	models := make([]Model, 0, len(result))
	for _, val := range result {
		var model Model
		if err := json.Unmarshal([]byte(val), &model); err != nil {
			continue
		}
		models = append(models, model)
	}
	return models, nil
}

func (p *ProviderCache) ReLoad(ctx context.Context, providers []Provider, models []Model) error {
	pipe := p.rdb.TxPipeline()

	pipe.Del(ctx, ProviderAllKey)

	providerFields := make(map[string]interface{})
	for _, provider := range providers {
		data, err := json.Marshal(provider)
		if err != nil {
			return err
		}
		field := p.getProviderField(provider)
		providerFields[field] = data
		pipe.Del(ctx, fmt.Sprintf(ProviderKey, provider.Id))
	}
	if len(providerFields) > 0 {
		pipe.HSet(ctx, ProviderAllKey, providerFields)
	}

	for _, model := range models {
		data, err := json.Marshal(model)
		if err != nil {
			return err
		}
		field := p.modelKey(model.Name, model.Id)
		pipe.HSet(ctx, fmt.Sprintf(ProviderKey, model.Pid), field, data)
	}

	_, err := pipe.Exec(ctx)
	return err
}

type Provider struct {
	Id     int64  `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	APIKey string `json:"api_key,omitempty"`
}

type Model struct {
	Id          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Pid         int64  `json:"pid,omitempty"`
	InputPrice  int64  `json:"input_price,omitempty"`
	OutputPrice int64  `json:"output_price,omitempty"`
	PriceMode   string `json:"price_mode,omitempty"`
}
