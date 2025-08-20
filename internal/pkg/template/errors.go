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
	ErrTemplateEmpty          = errors.New("模板内容为空")
	ErrTemplateTooLarge       = errors.New("模板大小超出限制")
	ErrTemplateRenderTimeout  = errors.New("模板渲染超时")
	ErrTemplateOutputTooLarge = errors.New("模板输出大小超出限制")

	// 变量相关错误
	ErrTemplateVariableInvalid = errors.New("模板变量无效")

	// 函数相关错误
	ErrTemplateFunctionInvalid = errors.New("模板函数无效")
)

// Error 模板错误的包装
type Error struct {
	Op       string // 操作名称
	Template string // 模板内容（可能被截断）
	Err      error  // 原始错误
}

func (e *Error) Error() string {
	if e.Template != "" {
		return fmt.Sprintf("template %s error: %v (template: %s)", e.Op, e.Err, e.truncateTemplate())
	}
	return fmt.Sprintf("template %s error: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// truncateTemplate 截断模板内容用于错误显示
func (e *Error) truncateTemplate() string {
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
	return &Error{
		Op:       op,
		Template: template,
		Err:      err,
	}
}
