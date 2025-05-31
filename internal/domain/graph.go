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
	"github.com/ecodeclub/ekit"
)

type Node struct {
	ID       int64
	GraphID  int64
	Type     string
	Status   string
	Metadata ekit.AnyValue // 用于存储扩展属性
}

type Edge struct {
	ID       int64
	GraphID  int64
	SourceID int64
	TargetID int64
	Metadata ekit.AnyValue // 用于存储边的扩展属性
}

type Graph struct {
	ID       int64
	Steps    []Node
	Edges    []Edge
	Metadata ekit.AnyValue
}
