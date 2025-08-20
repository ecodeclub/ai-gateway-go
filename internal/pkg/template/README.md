# Template 包

ai-gateway 的通用模板渲染包，为 InvocationConfig 提供强大的 Prompt 模板渲染功能。

## 📋 目录

- [概述](#概述)
- [设计思路](#设计思路)
- [核心功能](#核心功能)
- [快速开始](#快速开始)
- [使用示例](#使用示例)
- [API 文档](#api-文档)
- [扩展性设计](#扩展性设计)
- [安全配置](#安全配置)
- [最佳实践](#最佳实践)

## 概述

Template 包是为 ai-gateway 项目设计的通用模板渲染引擎，主要用于在调用 LLM 前渲染 Prompt 模板。它支持：

- **Go 模板语法**：完整支持 Go `text/template` 语法
- **可扩展上下文**：支持 `data`、`attr` 及自定义数据源
- **自定义函数**：丰富的内置函数和自定义函数支持
- **安全控制**：多层安全机制防止模板注入
- **高性能**：模板缓存和并发安全设计

## 设计思路

### 架构原则

1. **开闭原则**：通过接口设计，支持无限扩展新的数据源和函数
2. **单一职责**：每个组件专注于特定功能
3. **依赖倒置**：依赖抽象而非具体实现
4. **安全第一**：内置多层安全控制机制

### 核心组件

```
Template Package
├── TemplateRenderer     # 主渲染器接口
├── TemplateContext      # 可扩展上下文系统
├── ContextProvider      # 数据提供者接口
├── FuncRegistry         # 函数注册系统
└── SecurityConfig       # 安全配置
```

### 数据流

```
用户输入 → LLM解析 → JSON数据 → TemplateContext → 模板渲染 → 最终Prompt → LLM调用
```

## 核心功能

### 1. 基础模板渲染

支持标准的 Go 模板语法：

```go
// 简单变量访问
{{ .data.name }}                    // 访问数据
{{ .attr.environment }}             // 访问属性

// 嵌套访问
{{ .data.user.profile.email }}      // 深层嵌套

// 条件判断
{{ if .data.vip }}VIP用户{{ else }}普通用户{{ end }}

// 循环处理
{{ range .data.items }}
- {{ .name }}: {{ .price }}
{{ end }}
```

### 2. 强大的函数系统

#### 字符串处理
```go
{{ .data.name | upper }}            // 大写转换
{{ .data.description | truncate 100 }} // 文本截断
{{ .data.tags | join ", " }}        // 数组连接
```

#### 数学运算
```go
{{ add .data.price .data.tax }}     // 加法
{{ div .data.total .data.count }}   // 除法
```

#### 日期格式化
```go
{{ .data.created_at | formatDate "2006-01-02" }}
{{ now | formatDate "2006-01-02 15:04:05" }}
```

#### 条件和默认值
```go
{{ .data.optional | default "默认值" }}
{{ if gt .data.score 80 }}优秀{{ end }}
```

### 3. 可扩展的上下文系统

轻松添加新的数据源：

```go
// 添加用户信息
context.RegisterProvider(NewStaticProvider("user", userData))

// 添加环境变量
context.RegisterProvider(NewStaticProvider("env", envData))

// 添加动态计算
context.RegisterProvider(NewFunctionProvider("calc", calcFunc))

// 现在可以使用：{{ .user.name }} {{ .env.database_url }} {{ .calc.timestamp }}
```

## 快速开始

### 基础使用

```go
package main

import (
    "context"
    "fmt"
    "github.com/ecodeclub/ai-gateway-go/internal/pkg/template"
)

func main() {
    // 1. 准备数据
    data := map[string]any{
        "name": "张三",
        "age":  30,
        "skills": []string{"Go", "Python", "JavaScript"},
    }
    
    attr := map[string]any{
        "company": "某科技公司",
        "position": "Go开发工程师",
    }
    
    // 2. 便捷渲染（推荐）
    result, err := template.RenderPrompt(context.Background(),
        "为{{ .attr.company }}的{{ .attr.position }}职位生成{{ .data.name }}的简历",
        data, attr)
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println(result)
    // 输出：为某科技公司的Go开发工程师职位生成张三的简历
}
```

### 高级使用

```go
func advancedExample() {
    // 1. 创建自定义渲染器
    renderer := template.NewDefaultRenderer(template.DefaultSecurityConfig())
    
    // 2. 注册自定义函数
    renderer.RegisterFunction("reverse", func(s string) string {
        runes := []rune(s)
        for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
            runes[i], runes[j] = runes[j], runes[i]
        }
        return string(runes)
    })
    
    // 3. 创建扩展上下文
    context := template.NewTemplateContext(nil)
    context.RegisterProvider(template.NewDataProvider(data))
    context.RegisterProvider(template.NewAttrProvider(attr))
    
    // 4. 添加自定义数据源
    context.RegisterProvider(template.NewStaticProvider("env", map[string]any{
        "version": "1.0.0",
        "debug":   true,
    }))
    
    // 5. 渲染复杂模板
    complexTemplate := `
{{ .data.name | reverse | upper }}的技能：
{{ range .data.skills }}
- {{ . }}
{{ end }}
环境版本：{{ .env.version }}
调试模式：{{ if .env.debug }}开启{{ else }}关闭{{ end }}
`
    
    result, err := renderer.Render(context.Background(), complexTemplate, context)
    // 处理结果...
}
```

## 使用示例

### 1. 简历生成场景

```go
func resumeExample() {
    // LLM 解析的简历数据
    data := map[string]any{
        "name": "王五",
        "experience": []map[string]any{
            {
                "company":    "ABC公司",
                "position":   "Go开发工程师",
                "startDate":  "2020-01-01",
                "endDate":    "2023-12-31",
                "achievements": []string{
                    "设计高并发系统",
                    "优化数据库性能",
                },
            },
        },
        "skills": []map[string]any{
            {"name": "Go", "level": 9},
            {"name": "Python", "level": 7},
        },
    }
    
    // 配置属性
    attr := map[string]any{
        "targetCompany":  "某大型科技公司",
        "targetPosition": "资深Go开发工程师",
    }
    
    // 复杂模板
    template := `
针对{{ .attr.targetCompany }}的{{ .attr.targetPosition }}职位优化简历：

=== 个人信息 ===
姓名：{{ .data.name }}

=== 工作经验 ===
{{ range .data.experience }}
**{{ .position }}** - {{ .company }}
时间：{{ .startDate }} 至 {{ .endDate }}
主要成就：
{{ range .achievements }}- {{ . }}
{{ end }}
{{ end }}

=== 技能评估 ===
{{ range .data.skills }}{{ if gte .level 8 }}⭐{{ end }} {{ .name }}: {{ .level }}/10
{{ end }}
`
    
    result, err := template.RenderPrompt(context.Background(), template, data, attr)
    // 输出完整的简历...
}
```

### 2. 扩展性演示

```go
func extensibilityDemo() {
    // 创建可扩展上下文
    context := template.NewTemplateContext(nil)
    
    // 原始需求：data 和 attr
    context.RegisterProvider(template.NewDataProvider(map[string]any{
        "name": "测试用户",
    }))
    
    context.RegisterProvider(template.NewAttrProvider(map[string]any{
        "env": "production",
    }))
    
    // 扩展1：添加用户信息
    context.RegisterProvider(template.NewStaticProvider("user", map[string]any{
        "id":   123,
        "role": "admin",
    }))
    
    // 扩展2：添加环境变量
    context.RegisterProvider(template.NewStaticProvider("env", map[string]any{
        "database_url": "mysql://localhost:3306/db",
        "redis_url":    "redis://localhost:6379",
    }))
    
    // 扩展3：添加动态计算
    context.RegisterProvider(template.NewFunctionProvider("calc", 
        func(ctx context.Context, params map[string]any) (any, error) {
            return map[string]any{
                "timestamp": time.Now().Unix(),
                "random":    rand.Intn(1000),
            }, nil
        }))
    
    // 现在可以使用所有数据源
    template := `
用户：{{ .data.name }} (ID: {{ .user.id }})
角色：{{ .user.role }}
环境：{{ .attr.env }}
数据库：{{ .env.database_url }}
时间戳：{{ .calc.timestamp }}
`
    
    renderer := template.NewDefaultRenderer(nil)
    result, err := renderer.Render(context.Background(), template, context)
    // 演示完整的扩展能力...
}
```

## API 文档

### 核心接口

#### TemplateRenderer

```go
type TemplateRenderer interface {
    // 渲染模板
    Render(ctx context.Context, templateStr string, context *TemplateContext) (string, error)
    
    // 注册自定义函数
    RegisterFunction(name string, fn any) error
    
    // 注册上下文提供者
    RegisterProvider(provider ContextProvider) error
    
    // 清除缓存
    ClearCache()
}
```

#### ContextProvider

```go
type ContextProvider interface {
    // 提供者名称
    Name() string
    
    // 提供数据
    Provide(ctx context.Context, params map[string]any) (any, error)
}
```

### 便捷函数

```go
// 便捷的 Prompt 渲染
func RenderPrompt(ctx context.Context, promptTemplate string, data, attr map[string]any) (string, error)

// 便捷的 SystemPrompt 渲染  
func RenderSystemPrompt(ctx context.Context, systemPromptTemplate string, data, attr map[string]any) (string, error)

// 创建渲染上下文
func CreateRenderContext(data, attr map[string]any, config *SecurityConfig) *TemplateContext
```

### 内置提供者

```go
// 数据提供者（处理 {{ .data.* }}）
func NewDataProvider(data map[string]any) *DataProvider

// 属性提供者（处理 {{ .attr.* }}）
func NewAttrProvider(attributes map[string]any) *AttrProvider

// 静态提供者（自定义静态数据）
func NewStaticProvider(name string, data any) *StaticProvider

// 函数提供者（动态计算数据）
func NewFunctionProvider(name string, fn func(ctx context.Context, params map[string]any) (any, error)) *FunctionProvider
```

## 扩展性设计

### 1. 添加新的数据源

遵循开闭原则，无需修改现有代码：

```go
// 实现 ContextProvider 接口
type UserProvider struct {
    userID int64
}

func (p *UserProvider) Name() string {
    return "user"
}

func (p *UserProvider) Provide(ctx context.Context, params map[string]any) (any, error) {
    // 从数据库或其他服务获取用户数据
    user := getUserFromDB(p.userID)
    return map[string]any{
        "id":    user.ID,
        "name":  user.Name,
        "email": user.Email,
    }, nil
}

// 注册使用
context.RegisterProvider(&UserProvider{userID: 123})

// 模板中使用
// {{ .user.name }} {{ .user.email }}
```

### 2. 添加自定义函数

```go
// 注册简单函数
renderer.RegisterFunction("formatCurrency", func(amount float64, currency string) string {
    return fmt.Sprintf("%.2f %s", amount, currency)
})

// 注册带错误返回的函数
renderer.RegisterFunction("httpGet", func(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    return string(body), err
})

// 模板中使用
// {{ formatCurrency .data.price "CNY" }}
// {{ httpGet "https://api.example.com/data" }}
```

## 安全配置

### 配置选项

```go
type SecurityConfig struct {
    MaxTemplateSize  int           // 模板最大大小
    MaxOutputSize    int           // 输出最大大小
    RenderTimeout    time.Duration // 渲染超时
    AllowedFunctions []string      // 函数白名单
    DisableHTTP      bool          // 禁用HTTP函数
    FunctionTimeout  time.Duration // 函数执行超时
    MaxContextDepth  int           // 最大嵌套深度
    MaxLoopCount     int           // 最大循环次数
}
```

### 预设配置

```go
// 默认配置（推荐生产环境）
config := template.DefaultSecurityConfig()

// 严格配置（高安全要求）
config := template.StrictSecurityConfig()

// 自定义配置
config := &template.SecurityConfig{
    MaxTemplateSize: 5 * 1024,    // 5KB
    MaxOutputSize:   50 * 1024,   // 50KB
    RenderTimeout:   time.Second * 3,
    AllowedFunctions: []string{
        "upper", "lower", "truncate", "formatDate", "add", "sub",
    },
    DisableHTTP: true,
}
```

## 最佳实践

### 1. 性能优化

```go
// ✅ 推荐：复用渲染器实例
renderer := template.NewDefaultRenderer(config)

// ✅ 推荐：使用模板缓存
// 相同模板会被自动缓存，避免重复解析

// ✅ 推荐：批量注册函数
for name, fn := range customFunctions {
    renderer.RegisterFunction(name, fn)
}

// ❌ 避免：每次创建新渲染器
// renderer := template.NewDefaultRenderer(config) // 在循环中
```

### 2. 错误处理

```go
// ✅ 推荐：详细的错误处理
result, err := renderer.Render(ctx, template, context)
if err != nil {
    var templateErr *template.TemplateError
    if errors.As(err, &templateErr) {
        log.Printf("模板错误：%s，模板：%s", templateErr.Op, templateErr.Template)
    }
    return err
}

// ✅ 推荐：验证模板语法
if err := config.ValidateTemplateSize(template); err != nil {
    return fmt.Errorf("模板过大：%w", err)
}
```

### 3. 安全考虑

```go
// ✅ 推荐：使用严格配置处理用户输入
strictRenderer := template.NewDefaultRenderer(template.StrictSecurityConfig())

// ✅ 推荐：函数白名单
config.AllowedFunctions = []string{"upper", "lower", "formatDate"}

// ✅ 推荐：输入验证
if len(userTemplate) > maxTemplateSize {
    return errors.New("模板过大")
}

// ❌ 避免：直接使用用户输入作为模板
// 应该先验证和清理
```

### 4. 测试建议

```go
func TestTemplateRendering(t *testing.T) {
    // ✅ 推荐：测试正常情况
    result, err := template.RenderPrompt(ctx, "Hello {{ .data.name }}", 
        map[string]any{"name": "World"}, nil)
    assert.NoError(t, err)
    assert.Equal(t, "Hello World", result)
    
    // ✅ 推荐：测试边界情况
    _, err = template.RenderPrompt(ctx, "", nil, nil)
    assert.Error(t, err)
    
    // ✅ 推荐：测试安全限制
    largeTemplate := strings.Repeat("a", 100000)
    _, err = template.RenderPrompt(ctx, largeTemplate, nil, nil)
    assert.Error(t, err)
}
```

## 总结

Template 包为 ai-gateway 提供了强大、安全、可扩展的模板渲染能力。通过精心设计的架构，它不仅满足了当前 `data` 和 `attr` 的需求，还为未来的扩展需求提供了完美的支持。

关键优势：
- 🚀 **高性能**：模板缓存、并发安全
- 🔒 **安全可靠**：多层安全控制机制  
- 🔧 **易于扩展**：开闭原则设计
- 📝 **功能丰富**：完整的 Go 模板语法 + 自定义函数
- 🧪 **测试完备**：62% 覆盖率，全面的测试用例

适用场景：
- InvocationConfig 的 Prompt 渲染
- 系统通知模板
- 报告生成
- 任何需要动态文本生成的场景
