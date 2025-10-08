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

	"github.com/ecodeclub/ekit"
)

type ChatStreamRequest struct {
	Sn                 string
	Messages           []Message
	InvocationConfigID int64
	Uid                int64
	Key                string
	PreviousResponseID string
	CallID             string
}

type Chat struct {
	Sn       string
	Uid      int64
	Title    string
	Messages []Message
	Ctime    time.Time
	Utime    time.Time
}

type Message struct {
	ID               int64  `json:"id"`
	Role             string `json:"role"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoningContent"`
	Ctime            time.Time
	Utime            time.Time
}

type ChatResponse struct {
	Sn       string
	Response Message
	Metadata ekit.AnyValue
}
