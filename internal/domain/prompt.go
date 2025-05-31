// Copyright 2021 ecodeclub
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
	_ OwnerType = "personal"
	_ OwnerType = "organization"
)

type Prompt struct {
	ID          int64
	Name        string
	Owner       int64
	OwnerType   OwnerType
	Description string
	// 当前发布版本的 id
	ActiveVersion int64
	// prompt 所有的版本信息
	Versions []PromptVersion
	Ctime    time.Time
	Utime    time.Time
}

type PromptVersion struct {
	ID            int64
	Label         string
	Content       string
	SystemContent string
	Temperature   float32
	TopN          float32
	MaxTokens     int
	Status        uint8
	Ctime         time.Time
	Utime         time.Time
}
