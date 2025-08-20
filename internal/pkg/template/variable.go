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

// Variable 变量接口，简洁的名称+值设计
type Variable interface {
	// Name 返回变量名称，用于模板中访问 {{ .name.xxx }}
	Name() string
	// Value 返回变量值，支持错误返回
	Value() (any, error)
	// SetValue 设置变量值
	SetValue(value any) error
}

// StaticVariable 静态变量实现，使用泛型支持类型安全
type StaticVariable[T any] struct {
	name  string
	value T
}

// NewVariable 创建静态变量
func NewVariable[T any](name string, value T) *StaticVariable[T] {
	return &StaticVariable[T]{
		name:  name,
		value: value,
	}
}

// Name 返回变量名称
func (v *StaticVariable[T]) Name() string {
	return v.name
}

// Value 返回变量值
func (v *StaticVariable[T]) Value() (any, error) {
	return v.value, nil
}

// SetValue 更新变量值
func (v *StaticVariable[T]) SetValue(value any) error {
	if typedValue, ok := value.(T); ok {
		v.value = typedValue
		return nil
	}
	return ErrTemplateVariableInvalid // 类型不匹配
}

// VariableGetter 变量获取器接口，用于动态变量访问其他变量
type VariableGetter interface {
	Get(name string) (any, error)
}

// DynamicFunc 动态函数类型，支持访问其他变量并返回错误
type DynamicFunc func(getter VariableGetter) (any, error)

// DynamicVariable 动态变量，支持函数计算和变量间依赖
type DynamicVariable struct {
	name   string
	fn     DynamicFunc
	getter VariableGetter
}

// NewDynamicVariable 创建动态变量
func NewDynamicVariable(name string, fn DynamicFunc) *DynamicVariable {
	return &DynamicVariable{
		name: name,
		fn:   fn,
	}
}

// Name 返回变量名称
func (v *DynamicVariable) Name() string {
	return v.name
}

// Value 返回计算后的变量值
func (v *DynamicVariable) Value() (any, error) {
	if v.fn != nil {
		return v.fn(v.getter)
	}
	return nil, ErrTemplateVariableInvalid
}

// SetValue 动态变量不支持直接设置值
func (v *DynamicVariable) SetValue(_ any) error {
	return ErrTemplateVariableInvalid // 动态变量不支持设置值
}

// SetGetter 设置变量获取器，用于访问其他变量
func (v *DynamicVariable) SetGetter(getter VariableGetter) {
	v.getter = getter
}
