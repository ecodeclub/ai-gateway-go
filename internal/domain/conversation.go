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

const (
	UNKNOWN = iota
	USER
	ASSISTANT
	SYSTEM
	TOOL
)

type Conversation struct {
	Sn       string
	Uid      string
	Title    string
	Messages []Message
	Time     string
}
type Message struct {
	ID               int64
	CID              int64
	Role             int64
	Content          string
	ReasoningContent string
}

type ChatResponse struct {
	Sn       string
	Response Message
	Metadata ekit.AnyValue
}
