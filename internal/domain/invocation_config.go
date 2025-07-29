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

package domain

import (
	"time"
)

type OwnerType string

func (o OwnerType) String() string {
	return string(o)
}

const (
	OwnerTypeUser         OwnerType = "user"
	OwnerTypeOrganization OwnerType = "organization"
)

type InvocationConfig struct {
	ID          int64
	Name        string
	Biz         BizConfig
	Description string
	// 配置的所有的版本信息
	Versions []InvocationCfgVersion
	Ctime    time.Time
	Utime    time.Time
}

type InvocationCfgVersionStatus string

const (
	InvocationCfgVersionStatusDraft  InvocationCfgVersionStatus = "draft"
	InvocationCfgVersionStatusActive InvocationCfgVersionStatus = "active"
)

func (s InvocationCfgVersionStatus) String() string {
	return string(s)
}

type InvocationCfgVersion struct {
	ID int64
	// InvocationConfig 的 ID
	InvID int64
	Model Model
	// 版本号
	Version      string
	Prompt       string
	SystemPrompt string
	Temperature  float32
	TopP         float32
	MaxTokens    int
	Status       InvocationCfgVersionStatus
	Ctime        time.Time
	Utime        time.Time
}
