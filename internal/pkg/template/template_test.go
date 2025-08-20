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

package template_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/ecodeclub/ai-gateway-go/internal/pkg/template"
)

// TemplateTestSuite 模板测试套件
type TemplateTestSuite struct {
	suite.Suite
	render *template.DefaultRender
}

// SetupTest 设置测试
func (s *TemplateTestSuite) SetupTest() {
	// 创建渲染器
	s.render = template.NewDefaultRender(template.DefaultConfig())
}

func (s *TemplateTestSuite) newTemplateContext(t *testing.T) *template.Context {
	t.Helper()

	ctx := template.NewContext(t.Context())
	// 注册基础变量
	err := ctx.SetVariable(template.NewVariable("data", map[string]any{
		"name":      "张三",
		"age":       25,
		"skills":    []string{"Go", "Python", "Java"},
		"timestamp": int64(1640995200), // Unix时间戳：2022-01-01
		"address": map[string]any{
			"city":     "北京",
			"district": "海淀区",
		},
	}))
	require.NoError(t, err)

	err = ctx.SetVariable(template.NewVariable("attr", map[string]any{
		"company":    "TechCorp",
		"department": "技术部",
		"level":      "高级工程师",
	}))
	require.NoError(t, err)
	return ctx
}

// TestBasicRendering 测试基础渲染
func (s *TemplateTestSuite) TestBasicRendering() {
	t := s.T()

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "简单变量访问",
			template: "姓名：{{ .data.name }}，年龄：{{ .data.age }}",
			expected: "姓名：张三，年龄：25",
		},
		{
			name:     "嵌套对象访问",
			template: "地址：{{ .data.address.city }}{{ .data.address.district }}",
			expected: "地址：北京海淀区",
		},
		{
			name:     "属性访问",
			template: "公司：{{ .attr.company }}，部门：{{ .attr.department }}",
			expected: "公司：TechCorp，部门：技术部",
		},
		{
			name:     "混合访问",
			template: "{{ .data.name }}在{{ .attr.company }}担任{{ .attr.level }}",
			expected: "张三在TechCorp担任高级工程师",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.render.Render(s.newTemplateContext(t), tt.template)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuiltinFunctions 测试内置函数
func (s *TemplateTestSuite) TestBuiltinFunctions() {
	t := s.T()

	tests := []struct {
		name     string
		template string
		expected string
	}{
		// 字符串处理函数
		{
			name:     "字符串大写",
			template: "{{ .data.name | upper }}",
			expected: "张三",
		},
		{
			name:     "字符串小写",
			template: "{{ \"HELLO\" | lower }}",
			expected: "hello",
		},
		{
			name:     "字符串去空格",
			template: "{{ \"  test  \" | trim }}",
			expected: "test",
		},
		{
			name:     "字符串截断",
			template: "{{ truncate 3 \"长字符串测试\" }}",
			expected: "长字符...",
		},
		{
			name:     "HTML转义",
			template: "{{ \"<script>\" | escape }}",
			expected: "&lt;script&gt;",
		},
		{
			name:     "数组连接",
			template: "技能：{{ .data.skills | join \", \" }}",
			expected: "技能：Go, Python, Java",
		},
		{
			name:     "字符串分割",
			template: "{{ len (split \"a,b,c\" \",\") }}",
			expected: "3",
		},
		{
			name:     "标题格式",
			template: "{{ \"hello world\" | title }}",
			expected: "Hello World",
		},
		{
			name:     "字符串替换",
			template: "{{ replace \"hello\" \"hi\" \"hello world\" }}",
			expected: "hi world",
		},

		// 数学函数
		{
			name:     "加法",
			template: "明年年龄：{{ add .data.age 1 }}",
			expected: "明年年龄：26",
		},
		{
			name:     "减法",
			template: "{{ sub 10 3 }}",
			expected: "7",
		},
		{
			name:     "乘法",
			template: "{{ mul 6 7 }}",
			expected: "42",
		},
		{
			name:     "除法",
			template: "{{ div 10 2 }}",
			expected: "5",
		},
		{
			name:     "取模",
			template: "{{ mod 10 3 }}",
			expected: "1",
		},
		{
			name:     "四舍五入",
			template: "{{ round 3.6 }}",
			expected: "4",
		},
		{
			name:     "向上取整",
			template: "{{ ceil 3.1 }}",
			expected: "4",
		},
		{
			name:     "向下取整",
			template: "{{ floor 3.9 }}",
			expected: "3",
		},

		// 比较函数
		{
			name:     "等于",
			template: "{{ if eq .data.age 25 }}是25岁{{ end }}",
			expected: "是25岁",
		},
		{
			name:     "不等于",
			template: "{{ if ne .data.age 30 }}不是30岁{{ end }}",
			expected: "不是30岁",
		},
		{
			name:     "大于",
			template: "{{ if gt .data.age 18 }}成年人{{ else }}未成年{{ end }}",
			expected: "成年人",
		},
		{
			name:     "大于等于",
			template: "{{ if gte .data.age 25 }}25或以上{{ end }}",
			expected: "25或以上",
		},
		{
			name:     "小于",
			template: "{{ if lt .data.age 30 }}30以下{{ end }}",
			expected: "30以下",
		},
		{
			name:     "小于等于",
			template: "{{ if lte .data.age 25 }}25或以下{{ end }}",
			expected: "25或以下",
		},

		// 集合函数
		{
			name:     "长度",
			template: "技能数量：{{ len .data.skills }}",
			expected: "技能数量：3",
		},
		{
			name:     "包含检查",
			template: "{{ if contains .data.skills \"Go\" }}包含Go{{ end }}",
			expected: "包含Go",
		},

		// 条件和默认值
		{
			name:     "默认值",
			template: "{{ default \"默认值\" .data.missing }}",
			expected: "默认值",
		},
		{
			name:     "空值合并",
			template: "{{ coalesce .data.missing \"\" \"备用值\" }}",
			expected: "备用值",
		},

		// 类型转换
		{
			name:     "转字符串",
			template: "{{ toString .data.age }}岁",
			expected: "25岁",
		},
		{
			name:     "转整数",
			template: "{{ toInt \"42\" }}",
			expected: "42",
		},
		{
			name:     "转浮点数",
			template: "{{ toFloat \"3.14\" }}",
			expected: "3.14",
		},

		// 编码函数
		{
			name:     "Base64编码",
			template: "{{ base64Encode \"hello\" }}",
			expected: "aGVsbG8=",
		},
		{
			name:     "Base64解码",
			template: "{{ base64Decode \"aGVsbG8=\" }}",
			expected: "hello",
		},

		// 逻辑函数
		{
			name:     "逻辑与",
			template: "{{ if and (gt .data.age 18) (contains .data.skills \"Go\") }}成年且会Go{{ end }}",
			expected: "成年且会Go",
		},
		{
			name:     "逻辑或",
			template: "{{ if or (eq .data.age 25) (eq .data.age 30) }}25或30岁{{ end }}",
			expected: "25或30岁",
		},

		// JSON处理
		{
			name:     "JSON序列化",
			template: "{{ toJson .data.address }}",
			expected: "{\"city\":\"北京\",\"district\":\"海淀区\"}",
		},
		{
			name:     "JSON反序列化",
			template: "{{ (fromJson \"{\\\"test\\\":\\\"value\\\"}\").test }}",
			expected: "value",
		},

		// 日期函数
		{
			name:     "当前时间",
			template: "{{ formatDate \"2006\" (now) }}",
			expected: "202", // 部分匹配即可
		},
		{
			name:     "Unix时间戳格式化",
			template: "{{ formatDate \"2006-01-02\" .data.timestamp }}",
			expected: "2022-01-01",
		},

		// 其他条件测试
		{
			name:     "空值合并多个值",
			template: "{{ coalesce \"\" nil \"有值\" \"其他\" }}",
			expected: "有值",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.render.Render(s.newTemplateContext(t), tt.template)
			require.NoError(t, err)
			if tt.name == "当前时间" {
				assert.Contains(t, result, tt.expected) // 部分匹配
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}

	// 补充更多内置函数分支测试
	t.Run("补充函数分支测试", func(t *testing.T) {
		ctx := s.newTemplateContext(t)

		// 测试formatDate的不同分支
		tests := []struct {
			name     string
			template string
			wantErr  bool
		}{
			{
				name:     "formatDate字符串日期",
				template: "{{ formatDate \"2006-01-02\" \"2023-12-25\" }}",
				wantErr:  false,
			},
			{
				name:     "formatDate无效字符串",
				template: "{{ formatDate \"2006-01-02\" \"invalid-date\" }}",
				wantErr:  true,
			},
			{
				name:     "formatDate无效类型",
				template: "{{ formatDate \"2006-01-02\" 123.45 }}",
				wantErr:  true,
			},
			{
				name:     "contains字符串包含",
				template: "{{ contains \"hello world\" \"world\" }}",
				wantErr:  false,
			},
			{
				name:     "contains map包含",
				template: "{{ contains .data.address \"city\" }}",
				wantErr:  false,
			},
			{
				name:     "contains非字符串item",
				template: "{{ contains \"hello\" 123 }}",
				wantErr:  false,
			},
			{
				name:     "length非集合类型",
				template: "{{ len 123 }}",
				wantErr:  false,
			},
			{
				name:     "isEmpty测试指针",
				template: "{{ default \"空指针\" nil }}",
				wantErr:  false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := s.render.Render(ctx, tt.template)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

// TestCustomFunctions 测试自定义函数
func (s *TemplateTestSuite) TestCustomFunctions() {
	t := s.T()

	// 注册自定义函数
	templateContext := s.newTemplateContext(t)
	err := templateContext.SetFunction("greet", func(name string) string {
		return "你好，" + name + "！"
	})
	require.NoError(t, err)

	tpl := "{{ greet .data.name }}"
	result, err := s.render.Render(templateContext, tpl)
	require.NoError(t, err)
	assert.Equal(t, "你好，张三！", result)
}

// TestErrorHandling 测试错误处理
func (s *TemplateTestSuite) TestErrorHandling() {
	t := s.T()

	tests := []struct {
		name        string
		template    string
		ctx         func(t *testing.T) *template.Context
		assertError func(t assert.TestingT, err error, i ...interface{}) bool
	}{
		{
			name:        "无效模板语法",
			template:    "{{ .data.name",
			ctx:         s.newTemplateContext,
			assertError: assert.Error,
		},
		{
			name:        "访问不存在的字段",
			template:    "{{ .data.nonexistent }}",
			ctx:         s.newTemplateContext,
			assertError: assert.NoError, // Go模板会返回零值
		},
		{
			name:        "空模板",
			template:    "",
			ctx:         s.newTemplateContext,
			assertError: assert.Error,
		},
		{
			name:     "Variable.Value()返回错误",
			template: "值：{{ .error_var }}",
			ctx: func(t *testing.T) *template.Context {
				ctx := template.NewContext(context.Background())
				errorVar := template.NewDynamicVariable("error_var", func(getter template.VariableGetter) (any, error) {
					return nil, fmt.Errorf("变量计算失败")
				})
				err := ctx.SetVariable(errorVar)
				require.NoError(t, err)
				return ctx
			},
			assertError: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err) && assert.Contains(t, err.Error(), "变量计算失败")
			},
		},
		{
			name:     "StaticVariable类型不匹配",
			template: "{{ .int_var }}",
			ctx: func(t *testing.T) *template.Context {
				ctx := template.NewContext(context.Background())
				intVar := template.NewVariable("int_var", 42)
				err := ctx.SetVariable(intVar)
				require.NoError(t, err)

				// 尝试设置错误类型的值
				if variable, exists := ctx.Variable("int_var"); exists {
					if staticVar, ok := variable.(*template.StaticVariable[int]); ok {
						err := staticVar.SetValue("不是整数")
						assert.Error(t, err)
						assert.Contains(t, err.Error(), "模板变量无效")
					}
				}
				return ctx
			},
			assertError: assert.NoError, // 渲染本身不会失败，失败在SetValue
		},
		{
			name:     "DynamicVariable不支持SetValue",
			template: "{{ .dyn_var }}",
			ctx: func(t *testing.T) *template.Context {
				ctx := template.NewContext(context.Background())
				dynVar := template.NewDynamicVariable("dyn_var", func(getter template.VariableGetter) (any, error) {
					return "动态值", nil
				})
				err := ctx.SetVariable(dynVar)
				require.NoError(t, err)

				// 尝试设置动态变量的值（应该失败）
				if variable, exists := ctx.Variable("dyn_var"); exists {
					if dynVar, ok := variable.(*template.DynamicVariable); ok {
						err := dynVar.SetValue("新值")
						assert.Error(t, err)
						assert.Contains(t, err.Error(), "模板变量无效")
					}
				}
				return ctx
			},
			assertError: assert.NoError, // 渲染本身不会失败，失败在SetValue
		},
		{
			name:     "函数返回错误",
			template: "{{ errorFunc \"error\" }}",
			ctx: func(t *testing.T) *template.Context {
				ctx := template.NewContext(context.Background())
				err := ctx.SetFunction("errorFunc", func(input string) (string, error) {
					if input == "error" {
						return "", fmt.Errorf("函数执行失败：%s", input)
					}
					return "success", nil
				})
				require.NoError(t, err)
				return ctx
			},
			assertError: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err) && assert.Contains(t, err.Error(), "函数执行失败")
			},
		},
		{
			name:     "函数正常执行",
			template: "{{ errorFunc \"normal\" }}",
			ctx: func(t *testing.T) *template.Context {
				ctx := s.newTemplateContext(t)
				err := ctx.SetFunction("errorFunc", func(input string) (string, error) {
					if input == "error" {
						return "", fmt.Errorf("函数执行失败：%s", input)
					}
					return "success", nil
				})
				require.NoError(t, err)
				return ctx
			},
			assertError: assert.NoError,
		},
		{
			name:     "除零错误",
			template: "{{ div 10 0 }}",
			ctx:      s.newTemplateContext,
			assertError: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err) && assert.Contains(t, err.Error(), "除零错误")
			},
		},
		{
			name:        "类型转换错误",
			template:    "{{ toInt \"not_a_number\" }}",
			ctx:         s.newTemplateContext,
			assertError: assert.Error,
		},
		{
			name:        "Base64解码错误",
			template:    "{{ base64Decode \"invalid_base64!@#\" }}",
			ctx:         s.newTemplateContext,
			assertError: assert.Error,
		},
		{
			name:        "JSON解析错误",
			template:    "{{ fromJson \"{invalid json}\" }}",
			ctx:         s.newTemplateContext,
			assertError: assert.Error,
		},
		{
			name:     "DynamicVariable的fn为nil",
			template: "{{ .nil_func_var }}",
			ctx: func(t *testing.T) *template.Context {
				ctx := template.NewContext(context.Background())
				dynVar := template.NewDynamicVariable("nil_func_var", nil)
				err := ctx.SetVariable(dynVar)
				require.NoError(t, err)
				return ctx
			},
			assertError: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err) && assert.Contains(t, err.Error(), "模板变量无效")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.render.Render(tt.ctx(t), tt.template)
			tt.assertError(t, err)
		})
	}
}

// TestConfiguration 测试配置
func (s *TemplateTestSuite) TestConfiguration() {
	t := s.T()

	// 测试严格配置
	strictRenderer := template.NewDefaultRender(template.StrictConfig())

	// 测试模板大小限制
	largeTemplate := strings.Repeat("{{ .data.name }}", 1000)
	_, err := strictRenderer.Render(s.newTemplateContext(t), largeTemplate)
	assert.Error(t, err) // 应该超过严格配置的大小限制

	// 测试超时控制（简化模板）
	slowTemplate := "{{ range .data.skills }}{{ . }}{{ end }}"
	_, err = strictRenderer.Render(s.newTemplateContext(t), slowTemplate)
	assert.NoError(t, err) // 这个简单模板不应该超时
}

// TestRealWorldScenario 测试真实场景
func (s *TemplateTestSuite) TestRealWorldScenario() {
	t := s.T()

	// 模拟简历数据
	resumeData := map[string]any{
		"name":  "王五",
		"age":   28,
		"email": "wangwu@example.com",
		"experience": []map[string]any{
			{
				"company":  "ABC公司",
				"position": "前端工程师",
				"duration": "2020-2022",
			},
			{
				"company":  "XYZ公司",
				"position": "全栈工程师",
				"duration": "2022-现在",
			},
		},
		"skills": []string{"JavaScript", "React", "Node.js", "Go"},
	}

	// 模拟职位要求
	jobAttr := map[string]any{
		"position":     "高级工程师",
		"company":      "未来科技",
		"requirements": []string{"React", "Go", "微服务"},
	}

	// 创建上下文
	ctx := template.NewContext(t.Context())
	require.NoError(t, ctx.SetVariable(template.NewVariable("data", resumeData)))
	require.NoError(t, ctx.SetVariable(template.NewVariable("attr", jobAttr)))

	// LLM提示模板
	promptTemplate := `
根据以下候选人信息，生成一份简洁的面试评估：

候选人：{{ .data.name }}
年龄：{{ .data.age }}岁
邮箱：{{ .data.email }}

工作经历：
{{- range .data.experience }}
- {{ .company }}：{{ .position }}（{{ .duration }}）
{{- end }}

技能：{{ .data.skills | join "、" }}

目标职位：{{ .attr.position }}@{{ .attr.company }}
职位要求：{{ .attr.requirements | join "、" }}

评估建议：
1. 技能匹配度：{{ if contains .data.skills "React" }}✓ React{{ else }}✗ React{{ end }}，{{ if contains .data.skills "Go" }}✓ Go{{ else }}✗ Go{{ end }}
2. 经验水平：{{ len .data.experience }}段工作经历，适合{{ .attr.position }}职位
3. 推荐等级：{{ if and (contains .data.skills "React") (contains .data.skills "Go") }}强烈推荐{{ else }}一般推荐{{ end }}
`

	result, err := s.render.Render(ctx, promptTemplate)
	require.NoError(t, err)

	// 验证关键内容
	assert.Contains(t, result, "王五")
	assert.Contains(t, result, "28岁")
	assert.Contains(t, result, "ABC公司")
	assert.Contains(t, result, "✓ React")
	assert.Contains(t, result, "✓ Go")
	assert.Contains(t, result, "强烈推荐")
}

// TestDynamicVariable 测试动态数据生成和渲染
func (s *TemplateTestSuite) TestDynamicVariables() {
	t := s.T()

	// 1. 创建Context和Render
	ctx := template.NewContext(context.Background())
	render := template.NewDefaultRender(template.DefaultConfig())

	// 2. 注册静态数据（用户简历基础信息）
	resumeData := map[string]any{
		"name":             "李小明",
		"age":              28,
		"skills":           []string{"Go", "Python", "JavaScript", "Docker"},
		"experience_years": 5,
	}
	err := ctx.SetVariable(template.NewVariable("resume", resumeData))
	require.NoError(t, err)

	// 3. 注册动态变量（实时计算的数据）

	// 3.1 动态生成当前时间
	err = ctx.SetVariable(template.NewDynamicVariable("current_time", func(getter template.VariableGetter) (any, error) {
		return time.Now().Format("2006-01-02 15:04:05"), nil
	}))
	require.NoError(t, err)

	// 3.2 动态生成随机推荐分数
	err = ctx.SetVariable(template.NewDynamicVariable("recommendation_score", func(getter template.VariableGetter) (any, error) {
		return 85, nil // 固定值用于测试
	}))
	require.NoError(t, err)

	// 3.3 动态生成技能匹配度（基于职位要求）
	jobRequirements := []string{"Go", "Docker", "Kubernetes"}
	err = ctx.SetVariable(template.NewDynamicVariable("skill_match", func(getter template.VariableGetter) (any, error) {
		resume, err1 := getter.Get("resume")
		if err1 != nil {
			return nil, err1
		}
		if resumeMap, ok := resume.(map[string]any); ok {
			resumeSkills := resumeMap["skills"].([]string)
			matchCount := 0
			for _, req := range jobRequirements {
				for _, skill := range resumeSkills {
					if req == skill {
						matchCount++
						break
					}
				}
			}
			return map[string]any{
				"matched_skills": matchCount,
				"total_required": len(jobRequirements),
				"match_rate":     float64(matchCount) / float64(len(jobRequirements)) * 100,
			}, nil
		}
		return map[string]any{"matched_skills": 0, "total_required": len(jobRequirements), "match_rate": 0.0}, nil
	}))
	require.NoError(t, err)

	// 4. 注册自定义函数（动态生成数据的函数）

	// 4.1 函数：生成经验等级
	err = ctx.SetFunction("experienceLevel", func(years int) string {
		switch {
		case years >= 8:
			return "资深专家"
		case years >= 5:
			return "高级工程师"
		case years >= 2:
			return "中级工程师"
		default:
			return "初级工程师"
		}
	})
	require.NoError(t, err)

	// 5. 动态LLM提示模板
	promptTemplate := `候选人：{{ .resume.name }}，经验：{{ experienceLevel .resume.experience_years }}，技能匹配：{{ .skill_match.matched_skills }}/{{ .skill_match.total_required }}项`

	// 6. 渲染测试
	result, err := render.Render(ctx, promptTemplate)
	require.NoError(t, err)
	assert.Contains(t, result, "李小明")
	assert.Contains(t, result, "高级工程师")
	assert.Contains(t, result, "2/3项")
}

// TestAdvancedDynamicVariables 测试高级动态变量系统
func (s *TemplateTestSuite) TestAdvancedDynamicVariables() {
	t := s.T()

	// 1. 创建Context和Render
	ctx := template.NewContext(context.Background())
	render := template.NewDefaultRender(template.DefaultConfig())

	// 2. 注册静态数据（基础简历信息）
	resumeData := map[string]any{
		"name":             "张工程师",
		"age":              30,
		"base_salary":      20000,
		"experience_years": 7,
		"skills":           []string{"Go", "Python", "Docker", "Kubernetes", "React"},
		"projects": []map[string]any{
			{"name": "微服务架构", "complexity": 9},
			{"name": "容器化平台", "complexity": 8},
		},
	}
	err := ctx.SetVariable(template.NewVariable("resume", resumeData))
	require.NoError(t, err)

	// 3. 注册简单动态变量（无参数）
	err = ctx.SetVariable(template.NewDynamicVariable("random_factor", func(getter template.VariableGetter) (any, error) {
		return 0.75, nil // 固定值用于测试
	}))
	require.NoError(t, err)

	// 4. 注册复杂动态变量（依赖其他变量）

	// 4.1 技能评分（基于技能数量和经验）
	err = ctx.SetVariable(template.NewDynamicVariable("skill_score", func(getter template.VariableGetter) (any, error) {
		resume, err1 := getter.Get("resume")
		if err1 != nil {
			return nil, err1
		}
		if resumeMap, ok := resume.(map[string]any); ok {
			skills := resumeMap["skills"].([]string)
			experience := resumeMap["experience_years"].(int)

			// 技能数量 * 经验系数
			baseScore := len(skills) * 10
			experienceBonus := experience * 5

			return map[string]any{
				"base_score":       baseScore,
				"experience_bonus": experienceBonus,
				"total_score":      baseScore + experienceBonus,
				"skill_count":      len(skills),
			}, nil
		}
		return map[string]any{"total_score": 0}, nil
	}))
	require.NoError(t, err)

	// 4.2 综合评估（依赖技能评分和随机因子）
	err = ctx.SetVariable(template.NewDynamicVariable("comprehensive_evaluation", func(getter template.VariableGetter) (any, error) {
		skillScore, err := getter.Get("skill_score")
		if err != nil {
			return nil, err
		}
		randomFactor, err := getter.Get("random_factor")
		if err != nil {
			return nil, err
		}

		skillScoreMap := skillScore.(map[string]any)
		randomFactorValue := randomFactor.(float64)

		skillTotal := skillScoreMap["total_score"].(int)
		randomBonus := int(randomFactorValue * 100) // 0-100随机分
		finalScore := skillTotal + randomBonus

		// 评级
		var rating string
		switch {
		case finalScore >= 120:
			rating = "S级专家"
		case finalScore >= 100:
			rating = "A级高手"
		default:
			rating = "B级熟练"
		}

		return map[string]any{
			"skill_score":  skillTotal,
			"random_bonus": randomBonus,
			"final_score":  finalScore,
			"rating":       rating,
		}, nil
	}))
	require.NoError(t, err)

	// 5. 测试模板
	promptTemplate := `{{ .resume.name }}：{{ .comprehensive_evaluation.rating }}，总分{{ .comprehensive_evaluation.final_score }}`

	// 6. 渲染测试
	result, err := render.Render(ctx, promptTemplate)
	require.NoError(t, err)
	assert.Contains(t, result, "张工程师")
	assert.Contains(t, result, "级")

	// 7. 测试数据更新对依赖链的影响
	updatedResume := map[string]any{
		"name":             "张工程师",
		"age":              30,
		"base_salary":      20000,
		"experience_years": 8,                                                                       // 经验+1年
		"skills":           []string{"Go", "Python", "Docker", "Kubernetes", "React", "TypeScript"}, // 新增技能
		"projects": []map[string]any{
			{"name": "微服务架构", "complexity": 9},
			{"name": "容器化平台", "complexity": 8},
			{"name": "云原生改造", "complexity": 10}, // 新增项目
		},
	}

	// 更新变量值
	if resumeVar, exists := ctx.Variable("resume"); exists {
		if staticVar, ok := resumeVar.(*template.StaticVariable[map[string]any]); ok {
			err := staticVar.SetValue(updatedResume)
			require.NoError(t, err)
		}
	}

	// 重新渲染验证依赖链更新
	result2, err := render.Render(ctx, promptTemplate)
	require.NoError(t, err)
	assert.Contains(t, result2, "张工程师")
	assert.NotEqual(t, result, result2) // 结果应该不同，因为数据更新了
}

// TestSuite 运行测试套件
func TestSuite(t *testing.T) {
	suite.Run(t, new(TemplateTestSuite))
}
