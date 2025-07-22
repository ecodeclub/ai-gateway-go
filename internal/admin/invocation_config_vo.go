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

package admin

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ekit/slice"
)

type InvocationConfigVO struct {
	ID          int64                    `json:"id"`
	Name        string                   `json:"name"`
	BizID       int64                    `json:"bizID"`
	BizName     string                   `json:"bizName"`
	Versions    []InvocationCfgVersionVO `json:"versions"`
	Description string                   `json:"description"`
	Ctime       int64                    `json:"ctime"`
	Utime       int64                    `json:"utime"`
}

func (vo InvocationConfigVO) toDomain() domain.InvocationConfig {
	return domain.InvocationConfig{
		ID:   vo.ID,
		Name: vo.Name,
		Biz: domain.BizConfig{
			ID: vo.BizID,
		},
		Description: vo.Description,
		Versions: slice.Map(vo.Versions, func(idx int, src InvocationCfgVersionVO) domain.InvocationCfgVersion {
			return src.toDomain()
		}),
	}
}

type InvocationCfgVersionVO struct {
	ID    int64 `json:"id"`
	InvID int64 `json:"invID"`

	ModelID           int64  `json:"modelID"`
	ModelName         string `json:"modelName"`
	ModelProviderID   int64  `json:"modelProviderID"`
	ModelProviderName string `json:"modelProviderName"`

	Version      string  `json:"version"`
	Prompt       string  `json:"prompt"`
	SystemPrompt string  `json:"systemPrompt"`
	Temperature  float32 `json:"temperature"`
	TopP         float32 `json:"topP"`
	MaxTokens    int     `json:"maxTokens"`
	Status       string  `json:"status"`
	Ctime        int64   `json:"ctime"`
	Utime        int64   `json:"utime"`
}

func (vo InvocationCfgVersionVO) toDomain() domain.InvocationCfgVersion {
	return domain.InvocationCfgVersion{
		ID:           vo.ID,
		InvID:        vo.InvID,
		Prompt:       vo.Prompt,
		SystemPrompt: vo.SystemPrompt,
		Version:      vo.Version,
		Model:        domain.Model{ID: vo.ModelID},
		Temperature:  vo.Temperature,
		TopP:         vo.TopP,
		MaxTokens:    vo.MaxTokens,
		Status:       domain.InvocationCfgVersionStatus(vo.Status),
	}
}

func newInvocationVO(p domain.InvocationConfig) InvocationConfigVO {
	return InvocationConfigVO{
		ID:      p.ID,
		BizID:   p.Biz.ID,
		BizName: p.Biz.Name,
		Name:    p.Name,
		Versions: slice.Map(p.Versions, func(idx int, src domain.InvocationCfgVersion) InvocationCfgVersionVO {
			return newInvocationCfgVersion(src)
		}),
		Description: p.Description,
		Ctime:       p.Ctime.UnixMilli(),
		Utime:       p.Utime.UnixMilli(),
	}
}

func newInvocationCfgVersion(v domain.InvocationCfgVersion) InvocationCfgVersionVO {
	return InvocationCfgVersionVO{
		ID:                v.ID,
		InvID:             v.InvID,
		Prompt:            v.Prompt,
		Version:           v.Version,
		ModelID:           v.Model.ID,
		ModelName:         v.Model.Name,
		ModelProviderID:   v.Model.Provider.ID,
		ModelProviderName: v.Model.Provider.Name,
		SystemPrompt:      v.SystemPrompt,
		Temperature:       v.Temperature,
		TopP:              v.TopP,
		MaxTokens:         v.MaxTokens,
		Status:            v.Status.String(),
		Ctime:             v.Ctime.UnixMilli(),
		Utime:             v.Utime.UnixMilli(),
	}
}

type SaveInvocationConfigReq struct {
	Cfg InvocationConfigVO `json:"cfg"`
}

type ListInvocationConfigReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type SaveInvocationConfigVersionReq struct {
	Version InvocationCfgVersionVO `json:"version"`
}

type ListInvocationConfigVersionsReq struct {
	InvID  int64 `json:"invID"`
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
}
