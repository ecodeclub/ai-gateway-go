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

package ioc

import (
	"github.com/ecodeclub/ginx/session/cookie"
	"time"

	"github.com/ecodeclub/ginx/session"
	"github.com/ecodeclub/ginx/session/redis"
)

func InitSession() session.Provider {
	type Config struct {
		JwtKey     string        `yaml:"jwtKey"`
		Expiration time.Duration `yaml:"expiration"`
		Cookie     struct {
			Domain string `yaml:"domain"`
		} `json:"cookie"`
	}
	var cfg Config
	const day30 = time.Hour * 24 * 30
	provider := redis.NewSessionProvider(InitRedis(), cfg.JwtKey, day30)
	provider.TokenCarrier = &cookie.TokenCarrier{
		MaxAge:   int(day30.Seconds()),
		Name:     "ssid",
		Secure:   true,
		HttpOnly: true,
		Domain:   cfg.Cookie.Domain,
	}
	return provider
}
