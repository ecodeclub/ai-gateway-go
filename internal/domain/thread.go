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

// Thread 是一个抽象概念，它代表的是我在一个对话里面
// 为了处理用户的输入，而引入的不同处理方式/步骤等
type Thread struct {
	// CfgID 从设计上来说，应该不止一个 CfgID，但是目前只有一个
	CfgID        int64           `json:"cfgID,omitempty"`
	Conversation LLMConversation `json:"conversation,omitempty"`
}
