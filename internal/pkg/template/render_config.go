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
	"time"
)

// Config 模板渲染配置
type Config struct {
	// 模板限制
	MaxTemplateSize int           // 模板最大大小（字节）
	MaxOutputSize   int           // 输出最大大小（字节）
	RenderTimeout   time.Duration // 渲染超时时间

	// 函数限制
	AllowedFunctions []string      // 允许的函数白名单，空表示允许所有
	DisableHTTP      bool          // 是否禁用HTTP相关函数
	FunctionTimeout  time.Duration // 单个函数执行超时

	// 上下文限制
	MaxContextDepth int // 最大上下文嵌套深度
	MaxLoopCount    int // 最大循环次数
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxTemplateSize:  10 * 1024,       // 10KB
		MaxOutputSize:    100 * 1024,      // 100KB
		RenderTimeout:    time.Second * 5, // 5秒
		AllowedFunctions: nil,             // 允许所有函数
		DisableHTTP:      true,            // 默认禁用HTTP
		FunctionTimeout:  time.Second,     // 函数1秒超时
		MaxContextDepth:  10,              // 最大10层嵌套
		MaxLoopCount:     1000,            // 最大1000次循环
	}
}

// StrictConfig 返回严格配置
func StrictConfig() *Config {
	return &Config{
		MaxTemplateSize: 2 * 1024,        // 2KB
		MaxOutputSize:   10 * 1024,       // 10KB
		RenderTimeout:   time.Second * 2, // 2秒
		AllowedFunctions: []string{ // 只允许基础函数
			"upper", "lower", "trim", "truncate",
			"default", "formatDate", "add", "sub",
		},
		DisableHTTP:     true,                   // 禁用HTTP
		FunctionTimeout: time.Millisecond * 500, // 函数500ms超时
		MaxContextDepth: 5,                      // 最大5层嵌套
		MaxLoopCount:    100,                    // 最大100次循环
	}
}

// IsFunctionAllowed 检查函数是否被允许
func (c *Config) IsFunctionAllowed(funcName string) bool {
	if len(c.AllowedFunctions) == 0 {
		return true // 空白名单表示允许所有
	}

	for _, allowed := range c.AllowedFunctions {
		if allowed == funcName {
			return true
		}
	}
	return false
}

// ValidateTemplateSize 验证模板大小
func (c *Config) ValidateTemplateSize(template string) error {
	if len(template) == 0 {
		return ErrTemplateEmpty
	}
	if len(template) > c.MaxTemplateSize {
		return ErrTemplateTooLarge
	}
	return nil
}

// ValidateOutputSize 验证输出大小
func (c *Config) ValidateOutputSize(output string) error {
	if len(output) > c.MaxOutputSize {
		return ErrTemplateOutputTooLarge
	}
	return nil
}
