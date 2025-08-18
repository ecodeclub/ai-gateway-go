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

package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	cache "github.com/ecodeclub/ai-gateway-go/internal/repository/cache"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ekit/iox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ecodeclub/ai-gateway-go/internal/admin"
	testioc "github.com/ecodeclub/ai-gateway-go/internal/test/ioc"
	"github.com/stretchr/testify/suite"
)

type ProviderTestSuite struct {
	suite.Suite
	*testioc.TestApp
	dao   *dao.ProviderDao
	cache *cache.ProviderCache
}

func TestProviderTestSuite(t *testing.T) {
	suite.Run(t, new(ProviderTestSuite))
}

func (s *ProviderTestSuite) SetupSuite() {
	app := testioc.InitApp(testioc.TestOnly{})
	s.TestApp = app
	s.dao = dao.NewProviderDao(s.DB)
	s.cache = cache.NewProviderCache(s.Rdb)
}

func (s *ProviderTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE providers").Error
	if err != nil {
		fmt.Println(err)
	}
	// 清空 Redis 缓存
	s.Rdb.FlushDB(ctx)
}

func (s *ProviderTestSuite) TestProviderSave() {
	testCases := []struct {
		name   string
		before func(ctx context.Context, t *testing.T)
		after  func(ctx context.Context, t *testing.T)

		req     admin.ProviderVO
		wantRes Result[int64]
	}{
		{
			// 使用 ID = 1
			name: "新建",
			req: admin.ProviderVO{
				Name:   "模拟P",
				ApiKey: "123",
			},
			before: func(ctx context.Context, t *testing.T) {},
			after: func(ctx context.Context, t *testing.T) {
				p, err := s.dao.GetProvider(ctx, 1)
				assert.NoError(t, err)
				s.assertProvider(dao.Provider{
					ID:   1,
					Name: "模拟P",
				}, p, t)
				data, err := s.cache.GetProvider(ctx, 1)
				assert.NoError(t, err)
				assert.Equal(t, cache.Provider{
					ID:   1,
					Name: "模拟P",
					// 存储的是加密后的
					APIKey: p.APIKey,
				}, data)
			},
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 1,
			},
		},

		{
			// 使用 ID 2
			name: "更新",
			req: admin.ProviderVO{
				ID:     2,
				Name:   "模拟-2-new",
				ApiKey: "123-new",
			},
			before: func(ctx context.Context, t *testing.T) {
				_, err := s.dao.SaveProvider(ctx, dao.Provider{
					ID:     2,
					Name:   "模拟-2",
					APIKey: "123",
				})
				assert.NoError(t, err)
			},
			after: func(ctx context.Context, t *testing.T) {
				p, err := s.dao.GetProvider(ctx, 2)
				assert.NoError(t, err)
				s.assertProvider(dao.Provider{
					ID:   2,
					Name: "模拟-2-new",
				}, p, t)
				data, err := s.cache.GetProvider(ctx, 2)
				assert.NoError(t, err)
				assert.Equal(t, cache.Provider{
					ID:   2,
					Name: "模拟-2-new",
					// 存储的是加密后的
					APIKey: p.APIKey,
				}, data)
			},
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 2,
			},
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			tc.before(ctx, t)
			defer tc.after(ctx, t)
			req, err := http.NewRequest(http.MethodPost,
				"/provider/save", iox.NewJSONReader(tc.req))
			req.Header.Set("content-type", "application/json")
			require.NoError(t, err)
			recorder := NewJSONResponseRecorder[int64]()
			s.GinServer.ServeHTTP(recorder, req)
			val, err := recorder.Scan()
			assert.NoError(t, err)
			require.Equal(t, tc.wantRes, val)
		})
	}
}

func (s *ProviderTestSuite) assertProvider(expect, actual dao.Provider, t *testing.T) {
	assert.True(t, actual.Ctime > 0)
	actual.Ctime = 0
	assert.True(t, actual.Utime > 0)
	actual.Utime = 0
	assert.NotEmpty(t, actual.APIKey)
	actual.APIKey = ""
	assert.Equal(t, expect, actual)
}
