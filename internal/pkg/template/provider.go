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
	"fmt"
)

// Provider 上下文提供者接口
// 实现此接口可以向模板提供特定的数据源
type Provider interface {
	// Name 返回提供者名称，用于模板中访问 {{ .name.xxx }}
	Name() string
	// Provide 提供数据，params包含初始化参数
	Provide(ctx context.Context, params map[string]any) (any, error)
}

// VariableProvider 创建变量提供者，处理 {{ .data.xxx }} 语法
type VariableProvider[T any] struct {
	name string
	vals T
}

// NewVariableProvider 创建变量提供者
func NewVariableProvider[T any](name string, vals T) *VariableProvider[T] {
	return &VariableProvider[T]{name: name, vals: vals}
}

// Name 返回提供者名称
func (p *VariableProvider[T]) Name() string {
	return p.name
}

// Provide 提供数据
func (p *VariableProvider[T]) Provide(_ context.Context, params map[string]any) (any, error) {
	// 如果params中有新的数据则使用新的
	if newData, exists := params[p.name]; exists {
		if dataMap, ok := newData.(T); ok {
			return dataMap, nil
		}
		var zero T
		return nil, fmt.Errorf("变量%s必须是%T类型，当前是%T类型", p.name, zero, newData)
	}
	// 否则返回初始化时的
	return p.vals, nil
}

// FunctionProvider 函数提供者，可以提供动态计算的数据
type FunctionProvider struct {
	name string
	fn   func(ctx context.Context, params map[string]any) (any, error)
}

// NewFunctionProvider 创建函数提供者
func NewFunctionProvider(name string, fn func(ctx context.Context, params map[string]any) (any, error)) *FunctionProvider {
	return &FunctionProvider{
		name: name,
		fn:   fn,
	}
}

// Name 返回提供者名称
func (p *FunctionProvider) Name() string {
	return p.name
}

// Provide 通过函数提供数据
func (p *FunctionProvider) Provide(ctx context.Context, params map[string]any) (any, error) {
	if p.fn == nil {
		return nil, fmt.Errorf("function provider %s has no function", p.name)
	}
	return p.fn(ctx, params)
}
