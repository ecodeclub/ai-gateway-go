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

	"github.com/ecodeclub/ai-gateway-go/internal/admin"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	testioc "github.com/ecodeclub/ai-gateway-go/internal/test/ioc"
	"github.com/ecodeclub/ekit/iox"
	"github.com/ecodeclub/ginx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BizConfigTestSuite struct {
	suite.Suite
	*testioc.TestApp
	repo *repository.BizConfigRepository
}

func TestBizConfigTestSuite(t *testing.T) {
	suite.Run(t, new(BizConfigTestSuite))
}

func (s *BizConfigTestSuite) SetupSuite() {
	app := testioc.InitApp(testioc.TestOnly{})
	s.TestApp = app
	s.repo = repository.NewBizConfigRepository(dao.NewBizConfigDAO(s.DB))
}

func (s *BizConfigTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE biz_configs").Error
	s.NoError(err)
}

func (s *BizConfigTestSuite) TestConfig_Save() {
	t := s.T()

	testCases := []struct {
		name     string
		before   func(t *testing.T)
		req      admin.BizConfig
		after    func(t *testing.T, expected admin.BizConfig)
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name:   "新建成功",
			before: func(t *testing.T) {},
			req: admin.BizConfig{
				ID:        10001,
				Name:      "biz-name-10001",
				OwnerID:   1,
				OwnerType: "user",
				Config:    "config-10001",
			},
			after: func(t *testing.T, expected admin.BizConfig) {
				t.Helper()
				actual, err := s.repo.GetByID(t.Context(), 10001)
				require.NoError(t, err)
				s.assertConfig(t, expected, actual)
			},
			wantCode: 200,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 10001,
			},
		},
		{
			name: "更新成功",
			before: func(t *testing.T) {
				t.Helper()
				now := time.Now()
				_, err := s.repo.Save(t.Context(), domain.BizConfig{
					ID:        10002,
					Name:      "biz-name-10002",
					OwnerID:   2,
					OwnerType: "user",
					Config:    "config-10002",
					Ctime:     now,
					Utime:     now,
				})
				require.NoError(t, err)
			},
			req: admin.BizConfig{
				ID:        10002,
				Name:      "update-biz-name-10002",
				OwnerID:   3,
				OwnerType: "user",
				Config:    "update-biz-name-10002",
			},
			after: func(t *testing.T, expected admin.BizConfig) {
				t.Helper()
				actual, err := s.repo.GetByID(t.Context(), 10002)
				require.NoError(t, err)
				s.assertConfig(t, expected, actual)
			},
			wantCode: 200,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 10002,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t, tc.req)

			req, err := http.NewRequest(http.MethodPost,
				"/biz-configs/save", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[int64]()
			s.GinServer.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			require.Equal(t, tc.wantRes, recorder.MustScan())
		})
	}
}

func (s *BizConfigTestSuite) assertConfig(t *testing.T, expected admin.BizConfig, actual domain.BizConfig) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.OwnerID, actual.OwnerID)
	assert.Equal(t, expected.OwnerType, actual.OwnerType)
	assert.Equal(t, expected.Config, actual.Config)
	if expected.Ctime != 0 {
		assert.Equal(t, expected.Ctime, actual.Ctime)
	} else {
		assert.NotZero(t, actual.Ctime)
	}
	if expected.Utime != 0 {
		assert.Equal(t, expected.Utime, actual.Utime)
	} else {
		assert.NotZero(t, actual.Utime)
	}
}

func (s *BizConfigTestSuite) TestConfig_List() {
	t := s.T()

	err := s.DB.Exec("TRUNCATE TABLE `biz_configs`").Error
	require.NoError(t, err)

	total := 6
	expected := make([]admin.BizConfig, total)
	for i := range total {
		id := int64(10003 + i)
		_, err := s.repo.Save(t.Context(), domain.BizConfig{
			ID:        id,
			Name:      fmt.Sprintf("biz-name-%d", id),
			OwnerID:   id,
			OwnerType: "user",
			Config:    fmt.Sprintf("config-%d", id),
		})
		require.NoError(t, err)
		expected[total-i-1] = admin.BizConfig{
			ID:        id,
			Name:      fmt.Sprintf("biz-name-%d", id),
			OwnerID:   id,
			OwnerType: "user",
			Config:    fmt.Sprintf("config-%d", id),
		}
	}

	req := admin.ListReq{
		Offset: 0,
		Limit:  5,
	}

	httpReq, err := http.NewRequest(http.MethodPost,
		"/biz-configs/list", iox.NewJSONReader(req))
	require.NoError(t, err)
	httpReq.Header.Set("content-type", "application/json")
	recorder := NewJSONResponseRecorder[ginx.DataList[admin.BizConfig]]()
	s.GinServer.ServeHTTP(recorder, httpReq)
	require.Equal(t, 200, recorder.Code)
	result := recorder.MustScan()
	require.Equal(t, total, result.Data.Total)
	for i := range result.Data.List {
		require.NotZero(t, result.Data.List[i].Utime)
		require.NotZero(t, result.Data.List[i].Ctime)
		result.Data.List[i].Utime, result.Data.List[i].Ctime = 0, 0
		require.Equal(t, expected[i], result.Data.List[i])
	}
}

func (s *BizConfigTestSuite) TestConfig_Detail() {
	t := s.T()

	testCases := []struct {
		name             string
		before           func(t *testing.T)
		req              admin.IDReq
		wantCode         int
		assertResultFunc func(t *testing.T, r Result[admin.BizConfig])
	}{
		{
			name: "id存在",
			before: func(t *testing.T) {
				t.Helper()
				_, err := s.repo.Save(t.Context(), domain.BizConfig{
					ID:        10009,
					Name:      "biz-name-10009",
					OwnerID:   10009,
					OwnerType: "user",
					Config:    "config-10009",
				})
				require.NoError(t, err)
			},
			req: admin.IDReq{
				ID: 10009,
			},
			wantCode: 200,
			assertResultFunc: func(t *testing.T, r Result[admin.BizConfig]) {
				t.Helper()
				actual := r.Data
				require.NotZero(t, actual.Utime)
				require.NotZero(t, actual.Ctime)
				actual.Utime, actual.Ctime = 0, 0
				require.Equal(t, admin.BizConfig{
					ID:        10009,
					Name:      "biz-name-10009",
					OwnerID:   10009,
					OwnerType: "user",
					Config:    "config-10009",
				}, actual)
			},
		},
		{
			name:   "id不存在",
			before: func(t *testing.T) {},
			req: admin.IDReq{
				ID: -1,
			},
			wantCode: 500,
			assertResultFunc: func(t *testing.T, r Result[admin.BizConfig]) {
				t.Helper()
				require.Equal(t, Result[admin.BizConfig]{Code: 501001, Msg: "系统错误"}, r)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost,
				"/biz-configs/detail", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[admin.BizConfig]()
			s.GinServer.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			tc.assertResultFunc(t, recorder.MustScan())
		})
	}
}
