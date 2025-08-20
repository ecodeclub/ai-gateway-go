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

// DataProvider 数据提供者，处理 {{ .data.xxx }} 语法
// 用于提供LLM解析出的JSON结构化数据
type DataProvider struct {
	data map[string]any
}

// NewDataProvider 创建数据提供者
func NewDataProvider(data map[string]any) *DataProvider {
	if data == nil {
		data = make(map[string]any)
	}
	return &DataProvider{data: data}
}

// Name 返回提供者名称
func (p *DataProvider) Name() string {
	return "data"
}

// Provide 提供数据
func (p *DataProvider) Provide(ctx context.Context, params map[string]any) (any, error) {
	// 如果params中有新的data，则使用新的data
	if newData, exists := params["data"]; exists {
		if dataMap, ok := newData.(map[string]any); ok {
			return dataMap, nil
		}
		return nil, fmt.Errorf("data must be map[string]any, got %T", newData)
	}

	// 否则返回初始化时的data
	return p.data, nil
}

// SetData 更新数据
func (p *DataProvider) SetData(data map[string]any) {
	if data == nil {
		data = make(map[string]any)
	}
	p.data = data
}

// AttrProvider 属性提供者，处理 {{ .attr.xxx }} 语法
// 用于提供InvocationConfigVersion.Attributes中的配置数据
type AttrProvider struct {
	attributes map[string]any
}

// NewAttrProvider 创建属性提供者
func NewAttrProvider(attributes map[string]any) *AttrProvider {
	if attributes == nil {
		attributes = make(map[string]any)
	}
	return &AttrProvider{attributes: attributes}
}

// Name 返回提供者名称
func (p *AttrProvider) Name() string {
	return "attr"
}

// Provide 提供属性数据
func (p *AttrProvider) Provide(ctx context.Context, params map[string]any) (any, error) {
	// 如果params中有新的attr，则使用新的attr
	if newAttr, exists := params["attr"]; exists {
		if attrMap, ok := newAttr.(map[string]any); ok {
			return attrMap, nil
		}
		return nil, fmt.Errorf("attr must be map[string]any, got %T", newAttr)
	}

	// 否则返回初始化时的attributes
	return p.attributes, nil
}

// SetAttributes 更新属性
func (p *AttrProvider) SetAttributes(attributes map[string]any) {
	if attributes == nil {
		attributes = make(map[string]any)
	}
	p.attributes = attributes
}

// StaticProvider 静态数据提供者，用于提供固定的数据
// 可以用于扩展，比如用户信息、环境变量等
type StaticProvider struct {
	name string
	data any
}

// NewStaticProvider 创建静态数据提供者
func NewStaticProvider(name string, data any) *StaticProvider {
	return &StaticProvider{
		name: name,
		data: data,
	}
}

// Name 返回提供者名称
func (p *StaticProvider) Name() string {
	return p.name
}

// Provide 提供静态数据
func (p *StaticProvider) Provide(ctx context.Context, params map[string]any) (any, error) {
	// 检查params中是否有同名的数据，如果有则使用params中的
	if newData, exists := params[p.name]; exists {
		return newData, nil
	}

	// 否则返回静态数据
	return p.data, nil
}

// SetData 更新静态数据
func (p *StaticProvider) SetData(data any) {
	p.data = data
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
