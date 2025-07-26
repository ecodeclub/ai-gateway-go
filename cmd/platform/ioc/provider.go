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

package ioc

import (
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/gotomicro/ego/core/econf"
)

func initProvider(repo *repository.ProviderRepo) *service.ProviderService {
	type Config struct {
		Encrypt struct {
			Key string
		} `yaml:"encrypt"`
	}

	var cfg Config
	err := econf.UnmarshalKey("provider", &cfg)
	if err != nil {
		panic(err)
	}
	return service.NewProviderService(repo, cfg.Encrypt.Key)
}
