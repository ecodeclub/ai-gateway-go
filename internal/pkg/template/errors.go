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
	"errors"
	"fmt"
)

var (
	// 模板相关错误
	ErrTemplateEmpty    = errors.New("template is empty")
	ErrTemplateTooLarge = errors.New("template size exceeds limit")
	ErrTemplateInvalid  = errors.New("template syntax is invalid")
	ErrRenderTimeout    = errors.New("template render timeout")
	ErrOutputTooLarge   = errors.New("rendered output exceeds size limit")

	// 上下文相关错误
	ErrProviderNotFound = errors.New("context provider not found")
	ErrProviderInvalid  = errors.New("context provider is invalid")
	ErrDataNotFound     = errors.New("data not found in context")
	ErrDataTypeInvalid  = errors.New("data type is invalid")

	// 函数相关错误
	ErrFunctionNotFound   = errors.New("template function not found")
	ErrFunctionNotAllowed = errors.New("template function not allowed")
	ErrFunctionInvalid    = errors.New("template function is invalid")
	ErrFunctionTimeout    = errors.New("template function execution timeout")
)

// TemplateError 模板错误的包装
type TemplateError struct {
	Op       string // 操作名称
	Template string // 模板内容（可能被截断）
	Err      error  // 原始错误
}

func (e *TemplateError) Error() string {
	if e.Template != "" {
		return fmt.Sprintf("template %s error: %v (template: %s)", e.Op, e.Err, e.truncateTemplate())
	}
	return fmt.Sprintf("template %s error: %v", e.Op, e.Err)
}

func (e *TemplateError) Unwrap() error {
	return e.Err
}

// truncateTemplate 截断模板内容用于错误显示
func (e *TemplateError) truncateTemplate() string {
	if len(e.Template) <= 50 {
		return e.Template
	}
	return e.Template[:47] + "..."
}

// WrapError 包装错误为TemplateError
func WrapError(op, template string, err error) error {
	if err == nil {
		return nil
	}
	return &TemplateError{
		Op:       op,
		Template: template,
		Err:      err,
	}
}
