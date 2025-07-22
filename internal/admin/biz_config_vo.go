// Copyright 2023 ecodeclub
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

import "github.com/ecodeclub/ai-gateway-go/internal/domain"

type CreateBizConfigReq struct {
	ID        int64  `json:"id"`
	OwnerId   int64  `json:"owner_id"`
	OwnerType string `json:"owner_type"`
	Config    string `json:"config"`
}

type BizConfig struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	OwnerId   int64  `json:"ownerID"`
	OwnerType string `json:"ownerType"`
	Config    string `json:"config"`
}

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type BizConfigList struct {
	Cfgs  []BizConfig `json:"cfgs"`
	Total int
}

func newBizConfig(cfg domain.BizConfig) BizConfig {
	return BizConfig{
		ID:        cfg.ID,
		Name:      cfg.Name,
		OwnerId:   cfg.OwnerID,
		OwnerType: cfg.OwnerType,
		Config:    cfg.Config,
	}
}
