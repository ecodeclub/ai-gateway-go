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

package admin

import (
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ekit/slice"
)

type InvocationConfigVO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	BizID       int64  `json:"bizID"`
	BizName     string `json:"bizName,omitzero"`
	Description string `json:"description"`
	Ctime       int64  `json:"ctime"`
	Utime       int64  `json:"utime"`
}

func (vo InvocationConfigVO) toDomain() domain.InvocationConfig {
	return domain.InvocationConfig{
		ID:   vo.ID,
		Name: vo.Name,
		Biz: domain.BizConfig{
			ID: vo.BizID,
		},
		Description: vo.Description,
	}
}

type InvocationConfigVersionVO struct {
	ID    int64 `json:"id"`
	InvID int64 `json:"invID"`

	ModelID           int64  `json:"modelID"`
	ModelName         string `json:"modelName"`
	ModelProviderID   int64  `json:"modelProviderID"`
	ModelProviderName string `json:"modelProviderName"`

	Version      string         `json:"version"`
	Prompt       string         `json:"prompt"`
	SystemPrompt string         `json:"systemPrompt"`
	JSONSchema   string         `json:"jsonSchema"`
	Attributes   map[string]any `json:"attributes,omitempty"`
	Functions    []FunctionVO   `json:"functions,omitempty"`
	Temperature  float32        `json:"temperature"`
	TopP         float32        `json:"topP"`
	MaxTokens    int            `json:"maxTokens"`
	Status       string         `json:"status"`
	Ctime        int64          `json:"ctime"`
	Utime        int64          `json:"utime"`
}

type FunctionVO struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}

func (vo InvocationConfigVersionVO) toDomain() domain.InvocationConfigVersion {
	return domain.InvocationConfigVersion{
		ID:           vo.ID,
		Config:       domain.InvocationConfig{ID: vo.InvID},
		Model:        domain.Model{ID: vo.ModelID},
		Version:      vo.Version,
		Prompt:       vo.Prompt,
		SystemPrompt: vo.SystemPrompt,
		JSONSchema:   vo.JSONSchema,
		Attributes:   vo.Attributes,
		Functions: slice.Map(vo.Functions, func(_ int, src FunctionVO) domain.Function {
			return domain.Function{
				Name:       src.Name,
				Definition: src.Definition,
			}
		}),
		Temperature: vo.Temperature,
		TopP:        vo.TopP,
		MaxTokens:   vo.MaxTokens,
		Status:      domain.InvocationConfigVersionStatus(vo.Status),
		Ctime:       time.Time{},
		Utime:       time.Time{},
	}
}

func newInvocationVO(p domain.InvocationConfig) InvocationConfigVO {
	return InvocationConfigVO{
		ID:          p.ID,
		BizID:       p.Biz.ID,
		BizName:     p.Biz.Name,
		Name:        p.Name,
		Description: p.Description,
		Ctime:       p.Ctime.UnixMilli(),
		Utime:       p.Utime.UnixMilli(),
	}
}

func newInvocationCfgVersion(v domain.InvocationConfigVersion) InvocationConfigVersionVO {
	return InvocationConfigVersionVO{
		ID:                v.ID,
		InvID:             v.Config.ID,
		ModelID:           v.Model.ID,
		ModelName:         v.Model.Name,
		ModelProviderID:   v.Model.Provider.ID,
		ModelProviderName: v.Model.Provider.Name,
		Version:           v.Version,
		Prompt:            v.Prompt,
		SystemPrompt:      v.SystemPrompt,
		JSONSchema:        v.JSONSchema,
		Attributes:        v.Attributes,
		Functions: slice.Map(v.Functions, func(_ int, src domain.Function) FunctionVO {
			return FunctionVO{
				Name:       src.Name,
				Definition: src.Definition,
			}
		}),
		Temperature: v.Temperature,
		TopP:        v.TopP,
		MaxTokens:   v.MaxTokens,
		Status:      v.Status.String(),
		Ctime:       v.Ctime.UnixMilli(),
		Utime:       v.Utime.UnixMilli(),
	}
}

type ListInvocationConfigVersionsReq struct {
	InvID  int64 `json:"invID"`
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
}
