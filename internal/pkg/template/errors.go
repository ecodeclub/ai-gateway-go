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

// WrapError 包装错误并添加template包前缀和操作信息
func WrapError(op, template string, err error) error {
	// 包含包名的错误信息
	if template != "" {
		// 截断模板内容用于错误显示
		truncated := template
		if len(template) > 50 {
			truncated = template[:47] + "..."
		}
		return fmt.Errorf("template.%s: %w (模板: %s)", op, err, truncated)
	}
	return fmt.Errorf("template.%s: %w", op, err)
}
