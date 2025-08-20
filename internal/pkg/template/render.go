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
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"text/template"
)

// Render 模板渲染接口，简洁设计
type Render interface {
	// Render 渲染模板
	Render(ctx *Context, templateStr string) (string, error)
}

// DefaultRender 默认模板渲染器实现
type DefaultRender struct {
	mu        sync.RWMutex
	config    *Config                       // Config在渲染器层
	templates map[string]*template.Template // 模板缓存
}

// NewDefaultRender 创建默认渲染器
func NewDefaultRender(config *Config) *DefaultRender {
	if config == nil {
		config = DefaultConfig()
	}

	return &DefaultRender{
		config:    config,
		templates: make(map[string]*template.Template),
	}
}

// Render 渲染模板
func (r *DefaultRender) Render(ctx *Context, templateStr string) (string, error) {
	// 验证模板大小
	if err := r.config.ValidateTemplateSize(templateStr); err != nil {
		return "", WrapError("render", templateStr, err)
	}

	// 设置渲染超时
	renderCtx, cancel := context.WithTimeout(ctx, r.config.RenderTimeout)
	defer cancel()

	// 解析或获取缓存的模板
	tmpl, err := r.getOrParseTemplate(templateStr, ctx.FuncMap())
	if err != nil {
		return "", WrapError("render", templateStr, err)
	}

	// 构建模板数据
	data, err := ctx.VariableMap()
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

// getOrParseTemplate 获取或解析模板
func (r *DefaultRender) getOrParseTemplate(templateStr string, funcMap template.FuncMap) (*template.Template, error) {
	// 生成模板缓存键
	key := r.generateCacheKey(templateStr)

	// 尝试从缓存获取
	r.mu.RLock()
	if tmpl, exists := r.templates[key]; exists {
		r.mu.RUnlock()
		return tmpl, nil
	}
	r.mu.RUnlock()

	// 缓存未命中，需要解析模板
	r.mu.Lock()
	defer r.mu.Unlock()

	// 双重检查锁定
	if tmpl, exists := r.templates[key]; exists {
		return tmpl, nil
	}

	// 创建新模板
	tmpl := template.New("template")

	// 注册函数
	if funcMap != nil {
		tmpl = tmpl.Funcs(funcMap)
	}

	// 解析模板
	tmpl, err := tmpl.Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("模板解析失败: %w", err)
	}

	// 缓存模板
	r.templates[key] = tmpl

	return tmpl, nil
}

// generateCacheKey 生成缓存键
func (r *DefaultRender) generateCacheKey(templateStr string) string {
	hash := sha256.Sum256([]byte(templateStr))
	return fmt.Sprintf("%x", hash)
}

// executeTemplate 执行模板渲染
func (r *DefaultRender) executeTemplate(ctx context.Context, tmpl *template.Template, data map[string]any) (string, error) {
	// 创建缓冲区
	var buf strings.Builder

	// 使用goroutine执行模板渲染，以支持超时控制
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("模版执行panic: %v", r)
			}
		}()
		err := tmpl.Execute(&buf, data)
		done <- err
	}()

	// 等待完成或超时
	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("模板执行失败: %w", err)
		}
		return buf.String(), nil
	case <-ctx.Done():
		return "", ErrTemplateRenderTimeout
	}
}
