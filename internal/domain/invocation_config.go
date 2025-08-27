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
	"encoding/json"
	"github.com/tidwall/gjson"
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
	Ctime       time.Time
	Utime       time.Time
}

type InvocationConfigVersionStatus string

const (
	InvocationCfgVersionStatusDraft  InvocationConfigVersionStatus = "draft"
	InvocationCfgVersionStatusActive InvocationConfigVersionStatus = "active"
)

func (s InvocationConfigVersionStatus) String() string {
	return string(s)
}

func (s InvocationConfigVersionStatus) IsValid() bool {
	switch s {
	case InvocationCfgVersionStatusDraft, InvocationCfgVersionStatusActive:
		return true
	}
	return false
}

type InvocationConfigVersion struct {
	ID           int64
	Config       InvocationConfig
	Model        Model
	Version      string
	Prompt       string
	SystemPrompt string
	JSONSchema   string
	Attributes   Attributes
	Functions    []Function
	Temperature  float32
	TopP         float32
	MaxTokens    int
	Status       InvocationConfigVersionStatus
	Ctime        time.Time
	Utime        time.Time
}

type Attributes map[string]any

func (a Attributes) GetAttribute(expr string) map[string]any {
	if expr == "" {
		return a
	}
	attrJson := a.toJson()
	res := gjson.Get(attrJson, expr)
	attr := make(map[string]any)
	_ = json.Unmarshal([]byte(res.Raw), &attr)
	return attr
}

func (a Attributes) toJson() string {
	aByte, _ := json.Marshal(a)
	return string(aByte)
}

type Function struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}
