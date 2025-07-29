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

package test

import (
	"bytes"
	_ "embed"

	"github.com/gotomicro/ego/core/econf"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed config.yaml
	cfg string
)

func init() {
	// 用这种形式来规避运行部分测试加载配置失败的问题
	err := econf.LoadFromReader(bytes.NewReader([]byte(cfg)), yaml.Unmarshal)
	if err != nil {
		panic(err)
	}
}
