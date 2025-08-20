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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Context 模板上下文，内嵌标准Context，管理变量和函数
type Context struct {
	context.Context

	mu        sync.RWMutex
	variables map[string]Variable
	functions map[string]any
}

// NewContext 创建新的模板上下文
func NewContext(ctx context.Context) *Context {
	if ctx == nil {
		ctx = context.Background()
	}

	templateCtx := &Context{
		Context:   ctx,
		variables: make(map[string]Variable),
		functions: make(map[string]any),
	}

	// 注册内置函数
	templateCtx.registerBuiltinFunctions()
	return templateCtx
}

// SetVariable 注册变量
func (c *Context) SetVariable(variable Variable) error {
	if variable == nil {
		return ErrTemplateVariableInvalid
	}

	name := variable.Name()
	if name == "" {
		return ErrTemplateVariableInvalid
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.variables[name] = variable

	// 如果是动态变量，设置变量获取器
	if dynVar, ok := variable.(*DynamicVariable); ok {
		dynVar.SetGetter(c)
	}

	return nil
}

// Variable 获取指定名称的变量
func (c *Context) Variable(name string) (Variable, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	variable, exists := c.variables[name]
	return variable, exists
}

// VariableMap 构建模板执行的数据
func (c *Context) VariableMap() (map[string]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]any)

	// 构建变量数据
	for name, variable := range c.variables {
		value, err := variable.Value()
		if err != nil {
			return nil, WrapError("variable_map", "", err)
		}
		result[name] = value
	}

	return result, nil
}

// SetFunction 注册自定义函数
func (c *Context) SetFunction(name string, fn any) error {
	if err := c.validateFunction(fn); err != nil {
		return WrapError("set_function", "", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.functions[name] = fn
	return nil
}

// validateFunction 验证函数签名是否有效
func (c *Context) validateFunction(fn any) error {
	if fn == nil {
		return ErrTemplateFunctionInvalid
	}

	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return ErrTemplateFunctionInvalid
	}

	// 检查返回值数量（最多2个：value, error）
	fnType := v.Type()
	if fnType.NumOut() > 2 {
		return ErrTemplateFunctionInvalid
	}

	// 如果有2个返回值，第二个必须是error类型
	if fnType.NumOut() == 2 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if !fnType.Out(1).Implements(errorInterface) {
			return ErrTemplateFunctionInvalid
		}
	}

	return nil
}

// FuncMap 获取所有函数供模板使用
func (c *Context) FuncMap() template.FuncMap {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(template.FuncMap)
	for name, fn := range c.functions {
		result[name] = fn
	}
	return result
}

// Get 实现VariableGetter接口，用于动态变量访问其他变量
func (c *Context) Get(name string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if variable, exists := c.variables[name]; exists {
		return variable.Value()
	}
	return nil, nil // 变量不存在返回nil而不是错误
}

// registerBuiltinFunctions 注册所有内置函数
func (c *Context) registerBuiltinFunctions() {
	// 字符串处理函数
	c.functions["upper"] = strings.ToUpper
	c.functions["lower"] = strings.ToLower
	c.functions["trim"] = strings.TrimSpace
	c.functions["truncate"] = c.truncateString
	c.functions["escape"] = html.EscapeString
	c.functions["join"] = c.joinStrings
	c.functions["split"] = strings.Split
	c.functions["title"] = c.titleCase
	c.functions["replace"] = c.replaceString

	// 数学函数
	c.functions["add"] = c.mathAdd
	c.functions["sub"] = c.mathSub
	c.functions["mul"] = c.mathMul
	c.functions["div"] = c.mathDiv
	c.functions["mod"] = c.mathMod
	c.functions["round"] = math.Round
	c.functions["ceil"] = math.Ceil
	c.functions["floor"] = math.Floor

	// 比较函数
	c.functions["eq"] = c.equal
	c.functions["ne"] = c.notEqual
	c.functions["gt"] = c.greaterThan
	c.functions["gte"] = c.greaterThanOrEqual
	c.functions["lt"] = c.lessThan
	c.functions["lte"] = c.lessThanOrEqual

	// 日期格式化
	c.functions["formatDate"] = c.formatDate
	c.functions["now"] = time.Now
	c.functions["parseDate"] = c.parseDate

	// JSON处理
	c.functions["toJson"] = c.toJson
	c.functions["fromJson"] = c.fromJson

	// 条件和默认值
	c.functions["default"] = c.defaultValue
	c.functions["coalesce"] = c.coalesce

	// 编码函数
	c.functions["base64Encode"] = c.base64Encode
	c.functions["base64Decode"] = c.base64Decode

	// 类型转换
	c.functions["toString"] = c.toString
	c.functions["toInt"] = c.toInt
	c.functions["toFloat"] = c.toFloat

	// 集合函数
	c.functions["len"] = c.length
	c.functions["contains"] = c.contains

	// 逻辑函数
	c.functions["and"] = c.logicalAnd
	c.functions["or"] = c.logicalOr

	// 工具函数
	c.functions["isEmpty"] = c.isEmpty
}

// 内置函数实现
func (c *Context) truncateString(length int, text string) string {
	if length <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= length {
		return text
	}
	return string(runes[:length]) + "..."
}

func (c *Context) joinStrings(sep string, elems []string) string {
	return strings.Join(elems, sep)
}

func (c *Context) titleCase(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}

func (c *Context) replaceString(old, new, s string) string {
	return strings.ReplaceAll(s, old, new)
}

func (c *Context) mathAdd(a, b any) (float64, error) {
	aVal, err := c.toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := c.toFloat64(b)
	if err != nil {
		return 0, err
	}
	return aVal + bVal, nil
}

func (c *Context) mathSub(a, b any) (float64, error) {
	aVal, err := c.toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := c.toFloat64(b)
	if err != nil {
		return 0, err
	}
	return aVal - bVal, nil
}

func (c *Context) mathMul(a, b any) (float64, error) {
	aVal, err := c.toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := c.toFloat64(b)
	if err != nil {
		return 0, err
	}
	return aVal * bVal, nil
}

func (c *Context) mathDiv(a, b any) (float64, error) {
	aVal, err := c.toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := c.toFloat64(b)
	if err != nil {
		return 0, err
	}
	if bVal == 0 {
		return 0, fmt.Errorf("除零错误")
	}
	return aVal / bVal, nil
}

func (c *Context) mathMod(a, b any) (float64, error) {
	aVal, err := c.toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := c.toFloat64(b)
	if err != nil {
		return 0, err
	}
	if bVal == 0 {
		return 0, fmt.Errorf("取模零错误")
	}
	return math.Mod(aVal, bVal), nil
}

func (c *Context) equal(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

func (c *Context) notEqual(a, b any) bool {
	return !reflect.DeepEqual(a, b)
}

func (c *Context) greaterThan(a, b any) (bool, error) {
	aVal, err := c.toFloat64(a)
	if err != nil {
		return false, err
	}
	bVal, err := c.toFloat64(b)
	if err != nil {
		return false, err
	}
	return aVal > bVal, nil
}

func (c *Context) greaterThanOrEqual(a, b any) (bool, error) {
	aVal, err := c.toFloat64(a)
	if err != nil {
		return false, err
	}
	bVal, err := c.toFloat64(b)
	if err != nil {
		return false, err
	}
	return aVal >= bVal, nil
}

func (c *Context) lessThan(a, b any) (bool, error) {
	aVal, err := c.toFloat64(a)
	if err != nil {
		return false, err
	}
	bVal, err := c.toFloat64(b)
	if err != nil {
		return false, err
	}
	return aVal < bVal, nil
}

func (c *Context) lessThanOrEqual(a, b any) (bool, error) {
	aVal, err := c.toFloat64(a)
	if err != nil {
		return false, err
	}
	bVal, err := c.toFloat64(b)
	if err != nil {
		return false, err
	}
	return aVal <= bVal, nil
}

func (c *Context) formatDate(layout string, date any) (string, error) {
	switch v := date.(type) {
	case time.Time:
		return v.Format(layout), nil
	case string:
		layouts := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, l := range layouts {
			if t, err := time.Parse(l, v); err == nil {
				return t.Format(layout), nil
			}
		}
		return "", fmt.Errorf("无法解析日期字符串: %s", v)
	case int64:
		t := time.Unix(v, 0)
		return t.Format(layout), nil
	default:
		return "", fmt.Errorf("无效的日期类型: %T", date)
	}
}

func (c *Context) parseDate(dateStr, layout string) (time.Time, error) {
	return time.Parse(layout, dateStr)
}

func (c *Context) toJson(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Context) fromJson(jsonStr string) (any, error) {
	var result any
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

func (c *Context) defaultValue(value, defaultVal any) any {
	if c.isEmpty(value) {
		return defaultVal
	}
	return value
}

func (c *Context) coalesce(values ...any) any {
	for _, v := range values {
		if !c.isEmpty(v) {
			return v
		}
	}
	return nil
}

func (c *Context) base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func (c *Context) base64Decode(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func (c *Context) toString(v any) string {
	return fmt.Sprintf("%v", v)
}

func (c *Context) toInt(v any) (int64, error) {
	switch val := v.(type) {
	case int:
		return int64(val), nil
	case int64:
		return val, nil
	case float64:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	default:
		return 0, fmt.Errorf("无法将 %T 转换为整数", v)
	}
}

func (c *Context) toFloat(v any) (float64, error) {
	return c.toFloat64(v)
}

func (c *Context) length(v any) int {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return val.Len()
	default:
		return 0
	}
}

func (c *Context) contains(collection, item any) bool {
	val := reflect.ValueOf(collection)
	switch val.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			if reflect.DeepEqual(val.Index(i).Interface(), item) {
				return true
			}
		}
	case reflect.Map:
		return val.MapIndex(reflect.ValueOf(item)).IsValid()
	case reflect.String:
		if str, ok := item.(string); ok {
			return strings.Contains(val.String(), str)
		}
	}
	return false
}

func (c *Context) toFloat64(v any) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("无法将 %T 转换为浮点数", v)
	}
}

func (c *Context) isEmpty(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Array, reflect.Slice, reflect.Map:
		return val.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	default:
		return false
	}
}

func (c *Context) logicalAnd(a, b any) bool {
	return c.isTruthy(a) && c.isTruthy(b)
}

func (c *Context) logicalOr(a, b any) bool {
	return c.isTruthy(a) || c.isTruthy(b)
}

func (c *Context) isTruthy(v any) bool {
	if v == nil {
		return false
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Bool:
		return val.Bool()
	case reflect.String:
		return val.String() != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return val.Float() != 0
	case reflect.Array, reflect.Slice, reflect.Map:
		return val.Len() > 0
	case reflect.Ptr, reflect.Interface:
		return !val.IsNil()
	default:
		return true
	}
}
