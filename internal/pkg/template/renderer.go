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
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"text/template"
)

// Renderer 模板渲染器接口
type Renderer interface {
	// Render 渲染模板
	Render(ctx context.Context, templateStr string, context *Context) (string, error)

	// RegisterFunction 注册自定义函数
	RegisterFunction(name string, fn any) error

	// RegisterProvider 注册上下文提供者
	RegisterProvider(provider Provider) error

	// ClearCache 清除模板缓存
	ClearCache()
}

// DefaultTemplateRenderer 默认模板渲染器实现
type DefaultTemplateRenderer struct {
	mu        sync.RWMutex
	config    *SecurityConfig
	funcReg   *FunctionRegistry
	templates map[string]*template.Template // 模板缓存
	context   *Context
}

// NewDefaultRenderer 创建默认渲染器
func NewDefaultRenderer(config *SecurityConfig) *DefaultTemplateRenderer {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	renderer := &DefaultTemplateRenderer{
		config:    config,
		funcReg:   NewFuncRegistry(config),
		templates: make(map[string]*template.Template),
		context:   NewContext(config),
	}

	// 注册默认的data和attr提供者
	_ = renderer.context.RegisterProvider(NewVariableProvider[map[string]any]("data", nil))
	_ = renderer.context.RegisterProvider(NewVariableProvider[map[string]any]("attr", nil))

	return renderer
}

// Render 渲染模板
func (r *DefaultTemplateRenderer) Render(ctx context.Context, templateStr string, templateContext *Context) (string, error) {
	// 验证模板大小
	if err := r.config.ValidateTemplateSize(templateStr); err != nil {
		return "", WrapError("render", templateStr, err)
	}

	// 设置渲染超时
	renderCtx, cancel := context.WithTimeout(ctx, r.config.RenderTimeout)
	defer cancel()

	// 解析或获取缓存的模板
	tmpl, err := r.getOrParseTemplate(templateStr)
	if err != nil {
		return "", WrapError("render", templateStr, err)
	}

	// 构建模板上下文数据
	data, err := r.buildRenderData(renderCtx, templateContext)
	if err != nil {
		return "", WrapError("render", templateStr, err)
	}

	// 渲染模板
	result, err := r.executeTemplate(renderCtx, tmpl, data)
	if err != nil {
		return "", WrapError("render", templateStr, err)
	}

	// 验证输出大小
	if err := r.config.ValidateOutputSize(result); err != nil {
		return "", WrapError("render", templateStr, err)
	}

	return result, nil
}

// RegisterFunction 注册自定义函数
func (r *DefaultTemplateRenderer) RegisterFunction(name string, fn any) error {
	err := r.funcReg.Register(name, fn)
	if err != nil {
		return err
	}

	// 清除模板缓存，因为函数变更了
	r.ClearCache()
	return nil
}

// RegisterProvider 注册上下文提供者
func (r *DefaultTemplateRenderer) RegisterProvider(provider Provider) error {
	return r.context.RegisterProvider(provider)
}

// ClearCache 清除模板缓存
func (r *DefaultTemplateRenderer) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.templates = make(map[string]*template.Template)
	if r.context != nil {
		r.context.ClearAllCache()
	}
}

// getOrParseTemplate 获取或解析模板
func (r *DefaultTemplateRenderer) getOrParseTemplate(templateStr string) (*template.Template, error) {
	// 生成模板缓存键
	key := r.generateCacheKey(templateStr)

	r.mu.RLock()
	if tmpl, exists := r.templates[key]; exists {
		r.mu.RUnlock()
		return tmpl, nil
	}
	r.mu.RUnlock()

	// 解析模板
	tmpl, err := r.parseTemplate(templateStr)
	if err != nil {
		return nil, err
	}

	// 缓存模板
	r.mu.Lock()
	r.templates[key] = tmpl
	r.mu.Unlock()

	return tmpl, nil
}

// parseTemplate 解析模板
func (r *DefaultTemplateRenderer) parseTemplate(templateStr string) (*template.Template, error) {
	tmpl := template.New("").Funcs(r.funcReg.GetFuncMap())

	parsedTmpl, err := tmpl.Parse(templateStr)
	if err != nil {
		return nil, ErrTemplateInvalid
	}

	return parsedTmpl, nil
}

// buildRenderData 构建渲染数据
func (r *DefaultTemplateRenderer) buildRenderData(ctx context.Context, templateContext *Context) (map[string]any, error) {
	if templateContext == nil {
		// 使用默认上下文
		return r.context.BuildContext(ctx, nil)
	}

	// 使用提供的上下文
	return templateContext.BuildContext(ctx, nil)
}

// executeTemplate 执行模板渲染
func (r *DefaultTemplateRenderer) executeTemplate(ctx context.Context, tmpl *template.Template, data map[string]any) (string, error) {
	var buf bytes.Buffer

	// 创建一个可以取消的执行器
	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("template execution panic: %v", r)
			}
		}()

		err := tmpl.Execute(&buf, data)
		done <- err
	}()

	// 等待执行完成或超时
	select {
	case err := <-done:
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	case <-ctx.Done():
		return "", ErrRenderTimeout
	}
}

// generateCacheKey 生成缓存键
func (r *DefaultTemplateRenderer) generateCacheKey(templateStr string) string {
	hash := sha256.Sum256([]byte(templateStr))
	return fmt.Sprintf("%x", hash)
}

// CreateRenderContext 创建渲染上下文的便捷方法
func CreateRenderContext(data, attr map[string]any, config *SecurityConfig) *Context {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	ctx := NewContext(config)

	// 注册data和attr提供者
	if data != nil {
		_ = ctx.RegisterProvider(NewVariableProvider("data", data))
	} else {
		_ = ctx.RegisterProvider(NewVariableProvider("data", make(map[string]any)))
	}

	if attr != nil {
		_ = ctx.RegisterProvider(NewVariableProvider("attr", attr))
	} else {
		_ = ctx.RegisterProvider(NewVariableProvider("attr", make(map[string]any)))
	}

	return ctx
}

// Render 便捷的Prompt渲染方法
func Render(ctx context.Context, promptTemplate string, data, attr map[string]any) (string, error) {
	renderer := NewDefaultRenderer(DefaultSecurityConfig())
	renderContext := CreateRenderContext(data, attr, nil)
	return renderer.Render(ctx, promptTemplate, renderContext)
}
