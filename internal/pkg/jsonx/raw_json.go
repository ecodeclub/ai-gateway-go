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

package jsonx

import (
	"bytes"
	"encoding/json"

	"github.com/ecodeclub/ekit"
)

// RawJSON 一个代表 JSON 数据的类型
// 它可以是一个 JSON 数组，也可以是 JSON 对象
type RawJSON []byte

// Get 返回 key 对应的 value
func (c RawJSON) Get(name string) ekit.AnyValue {
	decoder := json.NewDecoder(bytes.NewReader(c))
	decoder.UseNumber()
	var res map[string]any
	err := decoder.Decode(&res)
	return ekit.AnyValue{Val: res[name], Err: err}
}
