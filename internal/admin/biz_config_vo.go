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

import "github.com/ecodeclub/ai-gateway-go/internal/domain"

type BizConfig struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	OwnerID   int64  `json:"ownerID"`
	OwnerType string `json:"ownerType"`
	Config    string `json:"config"`
	Ctime     int64  `json:"ctime"`
	Utime     int64  `json:"utime"`
}

func newBizConfig(cfg domain.BizConfig) BizConfig {
	return BizConfig{
		ID:        cfg.ID,
		Name:      cfg.Name,
		OwnerID:   cfg.OwnerID,
		OwnerType: cfg.OwnerType,
		Config:    cfg.Config,
		Ctime:     cfg.Ctime.UnixMilli(),
		Utime:     cfg.Utime.UnixMilli(),
	}
}
