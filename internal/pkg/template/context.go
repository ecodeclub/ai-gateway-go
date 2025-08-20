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

package template

import (
	"context"
	"sync"
)



// Context 模板上下文，支持动态注册和管理多个数据提供者
type Context struct {
	mu        sync.RWMutex
	providers map[string]Provider
	cache     map[string]any
	config    *SecurityConfig
}

// NewContext 创建新的模板上下文
func NewContext(config *SecurityConfig) *Context {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	return &Context{
		providers: make(map[string]Provider),
		cache:     make(map[string]any),
		config:    config,
	}
}

// RegisterProvider 注册上下文提供者
func (tc *Context) RegisterProvider(provider Provider) error {
	if provider == nil {
		return ErrProviderInvalid
	}

	name := provider.Name()
	if name == "" {
		return ErrProviderInvalid
	}

	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.providers[name] = provider
	// 清除缓存中对应的数据
	delete(tc.cache, name)

	return nil
}

// Provider 获取指定名称的提供者
func (tc *Context) Provider(name string) (Provider, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	provider, exists := tc.providers[name]
	return provider, exists
}

// BuildContext 构建模板执行的上下文数据
func (tc *Context) BuildContext(ctx context.Context, params map[string]any) (map[string]any, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	result := make(map[string]any)

	// 为每个注册的提供者生成数据
	for name, provider := range tc.providers {
		// 检查缓存
		if cached, exists := tc.cache[name]; exists {
			result[name] = cached
			continue
		}

		// 调用提供者生成数据
		data, err := provider.Provide(ctx, params)
		if err != nil {
			return nil, WrapError("build_context", "", err)
		}

		// 缓存结果
		tc.cache[name] = data
		result[name] = data
	}

	return result, nil
}

// ClearAllCache 清除所有缓存
func (tc *Context) ClearAllCache() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.cache = make(map[string]any)
}

// ClearProviderCache 清除指定提供者的缓存
func (tc *Context) ClearProviderCache(providerName string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	delete(tc.cache, providerName)
}

// ProviderNames 列出所有已注册的提供者名称
func (tc *Context) ProviderNames() []string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	names := make([]string, 0, len(tc.providers))
	for name := range tc.providers {
		names = append(names, name)
	}
	return names
}
