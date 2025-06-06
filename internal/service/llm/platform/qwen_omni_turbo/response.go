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

package qwen_omni_turbo

type StreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
			Audio   struct {
				Transcript string `json:"transcript"`
			} `json:"audio"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
		Index        int     `json:"index"`
	} `json:"choices"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	ID      string `json:"id"`
}
