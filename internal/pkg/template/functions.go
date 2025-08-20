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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TemplateFuncRegistry 模板函数注册器
type TemplateFuncRegistry struct {
	functions map[string]any
	config    *SecurityConfig
}

// NewFuncRegistry 创建函数注册器
func NewFuncRegistry(config *SecurityConfig) *TemplateFuncRegistry {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	registry := &TemplateFuncRegistry{
		functions: make(map[string]any),
		config:    config,
	}

	// 注册内置函数
	registry.registerBuiltinFunctions()
	return registry
}

// Register 注册自定义函数
func (r *TemplateFuncRegistry) Register(name string, fn any) error {
	if !r.config.IsFunctionAllowed(name) {
		return ErrFunctionNotAllowed
	}

	// 验证函数签名
	if err := r.validateFunction(fn); err != nil {
		return WrapError("register_function", "", err)
	}

	r.functions[name] = fn
	return nil
}

// GetFuncMap 获取所有函数供模板使用
func (r *TemplateFuncRegistry) GetFuncMap() template.FuncMap {
	result := make(template.FuncMap)
	for name, fn := range r.functions {
		if r.config.IsFunctionAllowed(name) {
			result[name] = fn
		}
	}
	return result
}

// validateFunction 验证函数签名是否有效
func (r *TemplateFuncRegistry) validateFunction(fn any) error {
	if fn == nil {
		return ErrFunctionInvalid
	}

	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return ErrFunctionInvalid
	}

	// 检查返回值数量（最多2个：value, error）
	fnType := v.Type()
	if fnType.NumOut() > 2 {
		return ErrFunctionInvalid
	}

	// 如果有2个返回值，第二个必须是error类型
	if fnType.NumOut() == 2 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if !fnType.Out(1).Implements(errorInterface) {
			return ErrFunctionInvalid
		}
	}

	return nil
}

// registerBuiltinFunctions 注册所有内置函数
func (r *TemplateFuncRegistry) registerBuiltinFunctions() {
	// 字符串处理函数
	r.functions["upper"] = strings.ToUpper
	r.functions["lower"] = strings.ToLower
	r.functions["trim"] = strings.TrimSpace
	r.functions["truncate"] = r.truncateString
	r.functions["escape"] = html.EscapeString
	r.functions["join"] = r.joinStrings
	r.functions["split"] = strings.Split
	r.functions["title"] = r.titleCase
	r.functions["replace"] = strings.ReplaceAll

	// 数学函数
	r.functions["add"] = r.mathAdd
	r.functions["sub"] = r.mathSub
	r.functions["mul"] = r.mathMul
	r.functions["div"] = r.mathDiv
	r.functions["mod"] = r.mathMod
	r.functions["round"] = math.Round
	r.functions["ceil"] = math.Ceil
	r.functions["floor"] = math.Floor

	// 比较函数
	r.functions["eq"] = r.equal
	r.functions["ne"] = r.notEqual
	r.functions["gt"] = r.greaterThan
	r.functions["gte"] = r.greaterThanOrEqual
	r.functions["lt"] = r.lessThan
	r.functions["lte"] = r.lessThanOrEqual

	// 日期格式化
	r.functions["formatDate"] = r.formatDate
	r.functions["now"] = time.Now
	r.functions["parseDate"] = r.parseDate

	// JSON处理
	r.functions["toJson"] = r.toJson
	r.functions["fromJson"] = r.fromJson

	// 条件和默认值
	r.functions["default"] = r.defaultValue
	r.functions["coalesce"] = r.coalesce

	// 编码函数
	r.functions["base64Encode"] = base64.StdEncoding.EncodeToString
	r.functions["base64Decode"] = r.base64Decode

	// 类型转换
	r.functions["toString"] = r.toString
	r.functions["toInt"] = r.toInt
	r.functions["toFloat"] = r.toFloat

	// 集合函数
	r.functions["len"] = r.length
	r.functions["contains"] = r.contains
}

// 字符串处理函数实现
func (r *TemplateFuncRegistry) truncateString(length int, text string) string {
	if length <= 0 {
		return ""
	}

	// 使用rune来正确处理Unicode字符（如中文）
	runes := []rune(text)
	if len(runes) <= length {
		return text
	}
	return string(runes[:length]) + "..."
}

func (r *TemplateFuncRegistry) joinStrings(sep string, elems []string) string {
	return strings.Join(elems, sep)
}

func (r *TemplateFuncRegistry) titleCase(s string) string {
	// 使用unicode-aware的title case
	caser := cases.Title(language.English)
	return caser.String(s)
}

// 数学函数实现
func (r *TemplateFuncRegistry) mathAdd(a, b any) (float64, error) {
	aVal, err := r.toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := r.toFloat64(b)
	if err != nil {
		return 0, err
	}
	return aVal + bVal, nil
}

func (r *TemplateFuncRegistry) mathSub(a, b any) (float64, error) {
	aVal, err := r.toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := r.toFloat64(b)
	if err != nil {
		return 0, err
	}
	return aVal - bVal, nil
}

func (r *TemplateFuncRegistry) mathMul(a, b any) (float64, error) {
	aVal, err := r.toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := r.toFloat64(b)
	if err != nil {
		return 0, err
	}
	return aVal * bVal, nil
}

func (r *TemplateFuncRegistry) mathDiv(a, b any) (float64, error) {
	aVal, err := r.toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := r.toFloat64(b)
	if err != nil {
		return 0, err
	}
	if bVal == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return aVal / bVal, nil
}

func (r *TemplateFuncRegistry) mathMod(a, b any) (float64, error) {
	aVal, err := r.toFloat64(a)
	if err != nil {
		return 0, err
	}
	bVal, err := r.toFloat64(b)
	if err != nil {
		return 0, err
	}
	if bVal == 0 {
		return 0, fmt.Errorf("modulo by zero")
	}
	return math.Mod(aVal, bVal), nil
}

// 比较函数实现
func (r *TemplateFuncRegistry) equal(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

func (r *TemplateFuncRegistry) notEqual(a, b any) bool {
	return !reflect.DeepEqual(a, b)
}

func (r *TemplateFuncRegistry) greaterThan(a, b any) (bool, error) {
	aVal, err := r.toFloat64(a)
	if err != nil {
		return false, err
	}
	bVal, err := r.toFloat64(b)
	if err != nil {
		return false, err
	}
	return aVal > bVal, nil
}

func (r *TemplateFuncRegistry) greaterThanOrEqual(a, b any) (bool, error) {
	aVal, err := r.toFloat64(a)
	if err != nil {
		return false, err
	}
	bVal, err := r.toFloat64(b)
	if err != nil {
		return false, err
	}
	return aVal >= bVal, nil
}

func (r *TemplateFuncRegistry) lessThan(a, b any) (bool, error) {
	aVal, err := r.toFloat64(a)
	if err != nil {
		return false, err
	}
	bVal, err := r.toFloat64(b)
	if err != nil {
		return false, err
	}
	return aVal < bVal, nil
}

func (r *TemplateFuncRegistry) lessThanOrEqual(a, b any) (bool, error) {
	aVal, err := r.toFloat64(a)
	if err != nil {
		return false, err
	}
	bVal, err := r.toFloat64(b)
	if err != nil {
		return false, err
	}
	return aVal <= bVal, nil
}

// 日期函数实现
func (r *TemplateFuncRegistry) formatDate(layout string, date any) (string, error) {
	switch v := date.(type) {
	case time.Time:
		return v.Format(layout), nil
	case string:
		// 尝试多种日期格式
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
		return "", fmt.Errorf("unable to parse date string: %s", v)
	case int64:
		t := time.Unix(v, 0)
		return t.Format(layout), nil
	default:
		return "", fmt.Errorf("invalid date type: %T", date)
	}
}

func (r *TemplateFuncRegistry) parseDate(dateStr, layout string) (time.Time, error) {
	return time.Parse(layout, dateStr)
}

// JSON函数实现
func (r *TemplateFuncRegistry) toJson(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *TemplateFuncRegistry) fromJson(jsonStr string) (any, error) {
	var result any
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

// 默认值函数实现
func (r *TemplateFuncRegistry) defaultValue(value, defaultVal any) any {
	if r.isEmpty(value) {
		return defaultVal
	}
	return value
}

func (r *TemplateFuncRegistry) coalesce(values ...any) any {
	for _, v := range values {
		if !r.isEmpty(v) {
			return v
		}
	}
	return nil
}

// 编码函数实现
func (r *TemplateFuncRegistry) base64Decode(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// 类型转换函数实现
func (r *TemplateFuncRegistry) toString(v any) string {
	return fmt.Sprintf("%v", v)
}

func (r *TemplateFuncRegistry) toInt(v any) (int64, error) {
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
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

func (r *TemplateFuncRegistry) toFloat(v any) (float64, error) {
	return r.toFloat64(v)
}

// 集合函数实现
func (r *TemplateFuncRegistry) length(v any) int {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return val.Len()
	default:
		return 0
	}
}

func (r *TemplateFuncRegistry) contains(collection, item any) bool {
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

// 辅助函数
func (r *TemplateFuncRegistry) toFloat64(v any) (float64, error) {
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
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func (r *TemplateFuncRegistry) isEmpty(v any) bool {
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
