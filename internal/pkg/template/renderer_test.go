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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TemplateTestSuite 模板测试套件
type TemplateTestSuite struct {
	suite.Suite
	renderer *DefaultTemplateRenderer
	ctx      context.Context
}

// SetupSuite 测试套件初始化
func (s *TemplateTestSuite) SetupSuite() {
	s.renderer = NewDefaultRenderer(DefaultSecurityConfig())
	s.ctx = context.Background()
}

// SetupTest 每个测试前的初始化
func (s *TemplateTestSuite) SetupTest() {
	// 清除缓存确保测试独立性
	s.renderer.ClearCache()
}

// TestBasicDataAccess 测试基础数据访问
func (s *TemplateTestSuite) TestBasicDataAccess() {
	t := s.T()

	data := map[string]any{
		"name":  "张三",
		"age":   30,
		"email": "zhangsan@example.com",
	}

	testCases := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "简单字符串访问",
			template: "Hello {{ .data.name }}",
			expected: "Hello 张三",
		},
		{
			name:     "数字访问",
			template: "Age: {{ .data.age }}",
			expected: "Age: 30",
		},
		{
			name:     "邮箱访问",
			template: "Email: {{ .data.email }}",
			expected: "Email: zhangsan@example.com",
		},
		{
			name:     "多字段组合",
			template: "{{ .data.name }} ({{ .data.age }}) - {{ .data.email }}",
			expected: "张三 (30) - zhangsan@example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderContext := CreateRenderContext(data, nil, nil)
			result, err := s.renderer.Render(s.ctx, tc.template, renderContext)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestBasicAttrAccess 测试基础属性访问
func (s *TemplateTestSuite) TestBasicAttrAccess() {
	t := s.T()

	attr := map[string]any{
		"environment": "production",
		"version":     "1.0.0",
		"company":     "某科技公司",
	}

	testCases := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "环境访问",
			template: "ENV: {{ .attr.environment }}",
			expected: "ENV: production",
		},
		{
			name:     "版本访问",
			template: "Version: {{ .attr.version }}",
			expected: "Version: 1.0.0",
		},
		{
			name:     "公司信息",
			template: "Company: {{ .attr.company }}",
			expected: "Company: 某科技公司",
		},
		{
			name:     "多属性组合",
			template: "{{ .attr.company }} - {{ .attr.environment }} - v{{ .attr.version }}",
			expected: "某科技公司 - production - v1.0.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderContext := CreateRenderContext(nil, attr, nil)
			result, err := s.renderer.Render(s.ctx, tc.template, renderContext)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestNestedDataAccess 测试嵌套数据访问
func (s *TemplateTestSuite) TestNestedDataAccess() {
	t := s.T()

	data := map[string]any{
		"user": map[string]any{
			"profile": map[string]any{
				"name":  "李四",
				"email": "lisi@example.com",
				"address": map[string]any{
					"city":    "北京",
					"country": "中国",
				},
			},
			"settings": map[string]any{
				"theme":    "dark",
				"language": "zh-CN",
			},
		},
		"skills": []map[string]any{
			{"name": "Go", "level": 8},
			{"name": "Python", "level": 7},
			{"name": "JavaScript", "level": 9},
		},
	}

	testCases := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "二级嵌套访问",
			template: "User: {{ .data.user.profile.name }}",
			expected: "User: 李四",
		},
		{
			name:     "三级嵌套访问",
			template: "City: {{ .data.user.profile.address.city }}",
			expected: "City: 北京",
		},
		{
			name:     "多路径访问",
			template: "{{ .data.user.profile.name }} from {{ .data.user.profile.address.country }}",
			expected: "李四 from 中国",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderContext := CreateRenderContext(data, nil, nil)
			result, err := s.renderer.Render(s.ctx, tc.template, renderContext)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestAdvancedTemplateFeatures 测试高级模板功能
func (s *TemplateTestSuite) TestAdvancedTemplateFeatures() {
	t := s.T()

	data := map[string]any{
		"vip":   true,
		"score": 85,
		"items": []map[string]any{
			{"name": "商品A", "price": 100},
			{"name": "商品B", "price": 200},
			{"name": "商品C", "price": 150},
		},
	}

	testCases := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "条件判断",
			template: "{{ if .data.vip }}VIP用户{{ else }}普通用户{{ end }}",
			expected: "VIP用户",
		},
		{
			name:     "数值比较",
			template: "{{ if gt .data.score 80 }}优秀{{ else }}良好{{ end }}",
			expected: "优秀",
		},
		{
			name: "循环处理",
			template: `商品列表：
{{ range .data.items }}- {{ .name }}: ¥{{ .price }}
{{ end }}`,
			expected: `商品列表：
- 商品A: ¥100
- 商品B: ¥200
- 商品C: ¥150
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderContext := CreateRenderContext(data, nil, nil)
			result, err := s.renderer.Render(s.ctx, tc.template, renderContext)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestCustomFunctions 测试自定义函数
func (s *TemplateTestSuite) TestCustomFunctions() {
	t := s.T()

	data := map[string]any{
		"name":        "JOHN DOE",
		"description": "这是一个非常长的描述文本，用来测试截断功能是否正常工作",
		"tags":        []string{"golang", "python", "javascript"},
		"price":       1234.56,
		"created_at":  "2024-01-01T10:30:00Z",
	}

	testCases := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "字符串转换",
			template: "Name: {{ .data.name | lower }}",
			expected: "Name: john doe",
		},
		{
			name:     "文本截断",
			template: "Desc: {{ .data.description | truncate 20 }}",
			expected: "Desc: 这是一个非常长的描述文本，用来测试截断功...",
		},
		{
			name:     "数组连接",
			template: "Tags: {{ .data.tags | join \", \" }}",
			expected: "Tags: golang, python, javascript",
		},
		{
			name:     "数学运算",
			template: "Total: {{ add .data.price 100 }}",
			expected: "Total: 1334.56",
		},
		{
			name:     "日期格式化",
			template: "Date: {{ .data.created_at | formatDate \"2006-01-02\" }}",
			expected: "Date: 2024-01-01",
		},
		{
			name:     "链式函数调用",
			template: "Name: {{ .data.name | lower | title }}",
			expected: "Name: John Doe",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderContext := CreateRenderContext(data, nil, nil)
			result, err := s.renderer.Render(s.ctx, tc.template, renderContext)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestExtensibilityDemo 测试扩展性演示
func (s *TemplateTestSuite) TestExtensibilityDemo() {
	t := s.T()

	// 创建自定义的渲染上下文
	renderContext := NewTemplateContext(DefaultSecurityConfig())

	// 注册原始的data和attr提供者
	_ = renderContext.RegisterProvider(NewDataProvider(map[string]any{
		"name": "张三",
		"age":  30,
	}))

	_ = renderContext.RegisterProvider(NewAttrProvider(map[string]any{
		"environment": "production",
		"version":     "1.0.0",
	}))

	// 演示添加新的变量类型 - 用户信息
	_ = renderContext.RegisterProvider(NewStaticProvider("user", map[string]any{
		"id":          123,
		"role":        "admin",
		"loginAt":     "2024-01-01 10:30:00",
		"permissions": []string{"read", "write", "admin"},
	}))

	// 演示添加新的变量类型 - 环境变量
	_ = renderContext.RegisterProvider(NewStaticProvider("env", map[string]any{
		"database_url": "mysql://localhost:3306/db",
		"redis_url":    "redis://localhost:6379",
		"debug":        true,
	}))

	// 演示添加函数提供者 - 动态计算数据
	_ = renderContext.RegisterProvider(NewFunctionProvider("calc", func(ctx context.Context, params map[string]any) (any, error) {
		return map[string]any{
			"timestamp": time.Now().Unix(),
			"random":    42,
			"status":    "active",
		}, nil
	}))

	testCases := []struct {
		name     string
		template string
		check    func(t *testing.T, result string)
	}{
		{
			name:     "原始data访问",
			template: "Name: {{ .data.name }}, Age: {{ .data.age }}",
			check: func(t *testing.T, result string) {
				assert.Equal(t, "Name: 张三, Age: 30", result)
			},
		},
		{
			name:     "原始attr访问",
			template: "ENV: {{ .attr.environment }}, Version: {{ .attr.version }}",
			check: func(t *testing.T, result string) {
				assert.Equal(t, "ENV: production, Version: 1.0.0", result)
			},
		},
		{
			name:     "新增用户信息访问",
			template: "User ID: {{ .user.id }}, Role: {{ .user.role }}",
			check: func(t *testing.T, result string) {
				assert.Equal(t, "User ID: 123, Role: admin", result)
			},
		},
		{
			name:     "新增环境变量访问",
			template: "DB: {{ .env.database_url }}, Debug: {{ .env.debug }}",
			check: func(t *testing.T, result string) {
				assert.Equal(t, "DB: mysql://localhost:3306/db, Debug: true", result)
			},
		},
		{
			name:     "动态计算数据访问",
			template: "Status: {{ .calc.status }}, Random: {{ .calc.random }}",
			check: func(t *testing.T, result string) {
				assert.Equal(t, "Status: active, Random: 42", result)
			},
		},
		{
			name: "多种数据源组合使用",
			template: `用户报告：
姓名：{{ .data.name }}（ID：{{ .user.id }}）
角色：{{ .user.role }}
环境：{{ .attr.environment }}
数据库：{{ .env.database_url }}
状态：{{ .calc.status }}`,
			check: func(t *testing.T, result string) {
				expected := `用户报告：
姓名：张三（ID：123）
角色：admin
环境：production
数据库：mysql://localhost:3306/db
状态：active`
				assert.Equal(t, expected, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := s.renderer.Render(s.ctx, tc.template, renderContext)
			require.NoError(t, err)
			tc.check(t, result)
		})
	}
}

// TestResumeScenario 测试简历修改实际场景
func (s *TemplateTestSuite) TestResumeScenario() {
	t := s.T()

	// 模拟LLM解析的简历数据
	data := map[string]any{
		"name":  "王五",
		"email": "wangwu@example.com",
		"phone": "13800138000",
		"experience": []map[string]any{
			{
				"company":     "ABC科技有限公司",
				"position":    "高级Go开发工程师",
				"startDate":   "2020-01-01",
				"endDate":     "2023-12-31",
				"description": "负责微服务架构设计和实现，主导了多个核心业务系统的开发，包括用户管理系统、订单处理系统等。技术栈包括Go、MySQL、Redis、Kubernetes等。",
				"achievements": []string{
					"设计并实现了高并发的用户认证系统，支持每秒10万次请求",
					"优化数据库查询性能，将关键接口响应时间从200ms降低到50ms",
					"建立了完整的CI/CD流程，提升了团队开发效率30%",
				},
			},
			{
				"company":     "DEF互联网公司",
				"position":    "Go开发工程师",
				"startDate":   "2018-06-01",
				"endDate":     "2019-12-31",
				"description": "参与电商平台后端开发，主要负责商品管理和库存管理模块。",
				"achievements": []string{
					"开发了分布式库存管理系统",
					"参与系统重构，提升系统稳定性",
				},
			},
		},
		"skills": []map[string]any{
			{"name": "Go", "level": 9, "years": 5},
			{"name": "Python", "level": 7, "years": 3},
			{"name": "MySQL", "level": 8, "years": 4},
			{"name": "Redis", "level": 8, "years": 3},
			{"name": "Kubernetes", "level": 7, "years": 2},
		},
		"education": map[string]any{
			"degree":     "计算机科学与技术学士",
			"university": "清华大学",
			"year":       "2018",
		},
		"expectedSalary": 35000,
		"available":      true,
	}

	// 配置属性
	attr := map[string]any{
		"targetCompany":  "某大型科技公司",
		"targetPosition": "资深Go开发工程师",
		"industry":       "互联网",
		"workLocation":   "北京",
		"currency":       "CNY",
		"template":       "professional",
	}

	// 复杂的简历模板
	template := `
基于以下信息为{{ .attr.targetCompany }}的{{ .attr.targetPosition }}职位生成专业简历：

=== 个人信息 ===
姓名：{{ .data.name }}
邮箱：{{ .data.email | lower }}
电话：{{ .data.phone }}
期望薪资：{{ .data.expectedSalary | toString }}{{ .attr.currency }}/月
工作地点：{{ .attr.workLocation }}
是否可立即到岗：{{ if .data.available }}是{{ else }}否{{ end }}

=== 教育背景 ===
{{ .data.education.degree }} - {{ .data.education.university }} ({{ .data.education.year }})

=== 工作经验 ===
{{ range .data.experience }}
**{{ .position }}** - {{ .company }}
时间：{{ .startDate }} 至 {{ .endDate }}
工作描述：{{ .description | truncate 100 }}

主要成就：
{{ range .achievements }}- {{ . }}
{{ end }}
{{ end }}

=== 技能清单 ===
{{ range .data.skills }}{{ if gte .level 8 }}⭐{{ end }} {{ .name }}：{{ .level }}/10 ({{ .years }}年经验)
{{ end }}

=== 职位匹配度分析 ===
目标职位：{{ .attr.targetPosition }}
所属行业：{{ .attr.industry }}
{{ if gt .data.expectedSalary 30000 }}
薪资水平：高级开发者水平
{{ else if gt .data.expectedSalary 20000 }}
薪资水平：中级开发者水平
{{ else }}
薪资水平：初级开发者水平
{{ end }}

=== 核心优势 ===
{{ $goSkill := "" }}
{{ range .data.skills }}{{ if eq .name "Go" }}{{ $goSkill = .level }}{{ end }}{{ end }}
{{ if gt $goSkill 8 }}
- Go语言专家级开发者，具备{{ range .data.skills }}{{ if eq .name "Go" }}{{ .years }}{{ end }}{{ end }}年实战经验
{{ end }}
- 丰富的{{ .attr.industry }}行业经验
- 具备大型项目架构设计和团队协作能力

此简历针对{{ .attr.targetCompany }}的{{ .attr.targetPosition }}职位进行了优化。
生成时间：{{ now | formatDate "2006-01-02 15:04:05" }}
`

	renderContext := CreateRenderContext(data, attr, nil)
	result, err := s.renderer.Render(s.ctx, template, renderContext)

	require.NoError(t, err)

	// 验证关键信息是否正确渲染
	assert.Contains(t, result, "王五")
	assert.Contains(t, result, "wangwu@example.com")
	assert.Contains(t, result, "某大型科技公司")
	assert.Contains(t, result, "资深Go开发工程师")
	assert.Contains(t, result, "35000CNY/月")
	assert.Contains(t, result, "ABC科技有限公司")
	assert.Contains(t, result, "高级Go开发工程师")
	assert.Contains(t, result, "⭐ Go：9/10")
	assert.Contains(t, result, "高级开发者水平")
	assert.Contains(t, result, "Go语言专家级开发者")

	// 打印完整结果以便人工检查
	fmt.Printf("=== 简历渲染结果 ===\n%s\n", result)
}

// TestErrorHandling 测试错误处理
func (s *TemplateTestSuite) TestErrorHandling() {
	t := s.T()

	testCases := []struct {
		name     string
		template string
		data     map[string]any
		attr     map[string]any
		wantErr  bool
		errType  error
	}{
		{
			name:     "空模板",
			template: "",
			wantErr:  true,
			errType:  ErrTemplateEmpty,
		},
		{
			name:     "模板语法错误",
			template: "{{ .data.name",
			wantErr:  true,
			errType:  ErrTemplateInvalid,
		},
		{
			name:     "访问不存在的数据",
			template: "{{ .data.nonexistent }}",
			data:     map[string]any{"name": "test"},
			wantErr:  false, // Go template默认返回零值
		},
		{
			name:     "除零错误",
			template: "{{ div 10 0 }}",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderContext := CreateRenderContext(tc.data, tc.attr, nil)
			result, err := s.renderer.Render(s.ctx, tc.template, renderContext)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorIs(t, err, tc.errType)
				}
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSecurityControls 测试安全控制
func (s *TemplateTestSuite) TestSecurityControls() {
	t := s.T()

	// 测试严格安全配置
	strictRenderer := NewDefaultRenderer(StrictSecurityConfig())

	testCases := []struct {
		name     string
		template string
		renderer *DefaultTemplateRenderer
		wantErr  bool
		errType  error
	}{
		{
			name:     "模板过大",
			template: strings.Repeat("a", 3000), // 超过严格配置的2KB限制
			renderer: strictRenderer,
			wantErr:  true,
			errType:  ErrTemplateTooLarge,
		},
		{
			name:     "正常大小模板",
			template: "{{ .data.name }}",
			renderer: strictRenderer,
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]any{"name": "test"}
			renderContext := CreateRenderContext(data, nil, nil)
			result, err := tc.renderer.Render(s.ctx, tc.template, renderContext)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorIs(t, err, tc.errType)
				}
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPerformanceAndConcurrency 测试性能和并发安全
func (s *TemplateTestSuite) TestPerformanceAndConcurrency() {
	t := s.T()

	template := "Hello {{ .data.name }}, your score is {{ .data.score }}"
	data := map[string]any{
		"name":  "Test User",
		"score": 95,
	}

	// 并发测试
	const concurrency = 100
	const iterations = 10

	results := make(chan error, concurrency*iterations)

	for i := 0; i < concurrency; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				renderContext := CreateRenderContext(data, nil, nil)
				result, err := s.renderer.Render(s.ctx, template, renderContext)

				if err != nil {
					results <- err
					return
				}

				expected := "Hello Test User, your score is 95"
				if result != expected {
					results <- fmt.Errorf("unexpected result: %s", result)
					return
				}
			}
			results <- nil
		}()
	}

	// 收集结果
	for i := 0; i < concurrency; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}

// TestConvenienceMethods 测试便捷方法
func (s *TemplateTestSuite) TestConvenienceMethods() {
	t := s.T()

	data := map[string]any{
		"name": "李华",
		"age":  25,
	}

	attr := map[string]any{
		"company": "示例公司",
		"role":    "开发工程师",
	}

	// 测试便捷的Prompt渲染方法
	promptResult, err := RenderPrompt(s.ctx, "Hello {{ .data.name }}, welcome to {{ .attr.company }}", data, attr)
	require.NoError(t, err)
	assert.Equal(t, "Hello 李华, welcome to 示例公司", promptResult)

	// 测试便捷的SystemPrompt渲染方法
	systemPromptResult, err := RenderSystemPrompt(s.ctx, "You are helping {{ .data.name }} with {{ .attr.role }} tasks", data, attr)
	require.NoError(t, err)
	assert.Equal(t, "You are helping 李华 with 开发工程师 tasks", systemPromptResult)
}

// TestRegisterFunction 测试自定义函数注册
func (s *TemplateTestSuite) TestRegisterFunction() {
	t := s.T()

	// 测试成功注册自定义函数
	customRenderer := NewDefaultRenderer(DefaultSecurityConfig())

	// 注册一个简单的自定义函数
	err := customRenderer.RegisterFunction("reverse", func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	})
	require.NoError(t, err)

	// 测试使用自定义函数
	renderContext := CreateRenderContext(map[string]any{"text": "hello"}, nil, nil)
	result, err := customRenderer.Render(s.ctx, "{{ .data.text | reverse }}", renderContext)
	require.NoError(t, err)
	assert.Equal(t, "olleh", result)

	// 测试注册复杂函数（带错误返回）
	err = customRenderer.RegisterFunction("safeDiv", func(a, b float64) (float64, error) {
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return a / b, nil
	})
	require.NoError(t, err)

	// 测试使用带错误返回的函数
	renderContext = CreateRenderContext(map[string]any{"a": 10.0, "b": 2.0}, nil, nil)
	result, err = customRenderer.Render(s.ctx, "{{ safeDiv .data.a .data.b }}", renderContext)
	require.NoError(t, err)
	assert.Equal(t, "5", result)

	// 测试注册无效函数
	err = customRenderer.RegisterFunction("invalid", "not a function")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFunctionInvalid)

	// 测试注册nil函数
	err = customRenderer.RegisterFunction("nil", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFunctionInvalid)

	// 测试注册返回值过多的函数
	err = customRenderer.RegisterFunction("tooManyReturns", func() (string, int, error) {
		return "", 0, nil
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFunctionInvalid)

	// 测试注册错误的返回值类型
	err = customRenderer.RegisterFunction("wrongErrorType", func() (string, string) {
		return "", ""
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFunctionInvalid)
}

// TestRegisterFunctionWithSecurity 测试带安全限制的函数注册
func (s *TemplateTestSuite) TestRegisterFunctionWithSecurity() {
	t := s.T()

	// 使用严格安全配置
	strictRenderer := NewDefaultRenderer(StrictSecurityConfig())

	// 尝试注册被禁止的函数
	err := strictRenderer.RegisterFunction("httpGet", func(url string) string {
		return "mock response"
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFunctionNotAllowed)

	// 注册允许的函数
	err = strictRenderer.RegisterFunction("upper", strings.ToUpper)
	assert.NoError(t, err)
}

// 运行测试套件
func TestTemplateTestSuite(t *testing.T) {
	suite.Run(t, new(TemplateTestSuite))
}
