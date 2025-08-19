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

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ekit/iox"
	"github.com/ecodeclub/ginx"
	"github.com/gotomicro/ego/core/econf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ecodeclub/ai-gateway-go/internal/admin"
	testioc "github.com/ecodeclub/ai-gateway-go/internal/test/ioc"
	"github.com/stretchr/testify/suite"
)

type ProviderTestSuite struct {
	suite.Suite
	*testioc.TestApp
	svc *service.ProviderService
}

func TestProviderTestSuite(t *testing.T) {
	suite.Run(t, new(ProviderTestSuite))
}

func (s *ProviderTestSuite) SetupSuite() {
	app := testioc.InitApp(testioc.TestOnly{})
	s.TestApp = app
	s.svc = service.NewProviderService(
		repository.NewProviderRepository(dao.NewProviderDAO(s.DB)),
		econf.GetString("provider.encrypt.key"))
}

func (s *ProviderTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE providers").Error
	s.NoError(err)
	err = s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE models").Error
	s.NoError(err)
}

func (s *ProviderTestSuite) TestProvider_Save() {
	t := s.T()

	testCases := []struct {
		name   string
		before func(t *testing.T)
		req    admin.ProviderVO
		after  func(t *testing.T, req admin.ProviderVO)

		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建成功",
			req: admin.ProviderVO{
				ID:     2001,
				Name:   "provider-2001",
				APIKey: "provider-apikey-2001",
			},
			before: func(t *testing.T) {},
			after: func(t *testing.T, req admin.ProviderVO) {
				t.Helper()
				p, err := s.svc.ProviderDetail(t.Context(), 2001)
				assert.NoError(t, err)
				s.assertProvider(t, req, p)
			},
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 2001,
			},
		},

		{
			name: "更新成功",
			before: func(t *testing.T) {
				t.Helper()
				_, err := s.svc.SaveProvider(t.Context(), domain.Provider{
					ID:     2002,
					Name:   "provider-2002",
					APIKey: "provider-apikey-2002",
				})
				assert.NoError(t, err)
			},
			req: admin.ProviderVO{
				ID:     2002,
				Name:   "provider-2002",
				APIKey: "provider-apikey-2002",
			},
			after: func(t *testing.T, req admin.ProviderVO) {
				t.Helper()
				p, err := s.svc.ProviderDetail(t.Context(), 2002)
				assert.NoError(t, err)
				s.assertProvider(t, req, p)
			},
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 2002,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)

			req, err := http.NewRequest(http.MethodPost,
				"/providers/save", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[int64]()
			s.GinServer.ServeHTTP(recorder, req)
			result := recorder.MustScan()
			assert.Equal(t, tc.wantCode, result.Code)
			require.Equal(t, tc.wantRes, result)

			tc.after(t, tc.req)
		})
	}
}

func (s *ProviderTestSuite) assertProvider(t *testing.T, expected admin.ProviderVO, actual domain.Provider) {
	t.Helper()
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.APIKey, actual.APIKey)
	if expected.Ctime != 0 {
		assert.Equal(t, expected.Ctime, actual.Ctime)
	} else {
		assert.True(t, actual.Ctime > 0)
	}
	if expected.Utime != 0 {
		assert.Equal(t, expected.Utime, actual.Utime)
	} else {
		assert.True(t, actual.Utime > 0)
	}
}

func (s *ProviderTestSuite) TestProvider_List() {
	t := s.T()

	err := s.DB.Exec("TRUNCATE TABLE `providers`").Error
	require.NoError(t, err)

	total := 6
	expected := make([]admin.ProviderVO, total)
	for i := range total {
		id := int64(3 + i)
		_, err := s.svc.SaveProvider(t.Context(), domain.Provider{
			ID:     id,
			Name:   fmt.Sprintf("provider-%d", id),
			APIKey: fmt.Sprintf("provider-apikey-%d", id),
		})
		require.NoError(t, err)
		expected[total-i-1] = admin.ProviderVO{
			ID:     id,
			Name:   fmt.Sprintf("provider-%d", id),
			APIKey: fmt.Sprintf("provider-apikey-%d", id),
		}
	}

	req := admin.ListReq{
		Offset: 0,
		Limit:  5,
	}

	httpReq, err := http.NewRequest(http.MethodPost,
		"/providers/list", iox.NewJSONReader(req))
	require.NoError(t, err)
	httpReq.Header.Set("content-type", "application/json")
	recorder := NewJSONResponseRecorder[ginx.DataList[admin.ProviderVO]]()
	s.GinServer.ServeHTTP(recorder, httpReq)
	require.Equal(t, 200, recorder.Code)
	result := recorder.MustScan()
	require.Equal(t, total, result.Data.Total)
	for i := range result.Data.List {
		require.True(t, result.Data.List[i].Utime > 0)
		require.True(t, result.Data.List[i].Ctime > 0)
		result.Data.List[i].Utime, result.Data.List[i].Ctime = 0, 0
		require.Equal(t, expected[i], result.Data.List[i])
	}
}

func (s *ProviderTestSuite) TestProvider_Detail() {
	t := s.T()

	testCases := []struct {
		name             string
		before           func(t *testing.T)
		req              admin.IDReq
		wantCode         int
		assertResultFunc func(t *testing.T, r Result[admin.ProviderVO])
	}{
		{
			name: "id存在_没有关联的模型",
			before: func(t *testing.T) {
				t.Helper()
				_, err := s.svc.SaveProvider(t.Context(), domain.Provider{
					ID:     2009,
					Name:   "provider-2009",
					APIKey: "provider-apikey-2009",
					Models: make([]domain.Model, 0),
				})
				require.NoError(t, err)
			},
			req: admin.IDReq{
				ID: 2009,
			},
			wantCode: 200,
			assertResultFunc: func(t *testing.T, r Result[admin.ProviderVO]) {
				t.Helper()
				actual := r.Data
				require.True(t, actual.Ctime > 0)
				actual.Ctime = 0
				require.True(t, actual.Utime > 0)
				actual.Utime = 0
				require.Equal(t, admin.ProviderVO{
					ID:     2009,
					Name:   "provider-2009",
					APIKey: "provider-apikey-2009",
				}, actual)
			},
		},
		{
			name: "id存在_有多个关联模型",
			before: func(t *testing.T) {
				t.Helper()
				pid, err := s.svc.SaveProvider(t.Context(), domain.Provider{
					ID:     2010,
					Name:   "provider-2010",
					APIKey: "provider-apikey-2010",
				})
				require.NoError(t, err)

				_, err = s.svc.SaveModel(t.Context(), domain.Model{
					ID:          2001,
					Provider:    domain.Provider{ID: pid},
					Name:        "model-2001",
					InputPrice:  2001,
					OutputPrice: 2001,
					PriceMode:   "2001",
				})
				require.NoError(t, err)

				_, err = s.svc.SaveModel(t.Context(), domain.Model{
					ID:          2002,
					Provider:    domain.Provider{ID: pid},
					Name:        "model-2002",
					InputPrice:  2002,
					OutputPrice: 2002,
					PriceMode:   "2002",
				})
				require.NoError(t, err)
			},
			req: admin.IDReq{
				ID: 2010,
			},
			wantCode: 200,
			assertResultFunc: func(t *testing.T, r Result[admin.ProviderVO]) {
				t.Helper()
				actual := r.Data
				require.True(t, actual.Ctime > 0)
				actual.Ctime = 0
				require.True(t, actual.Utime > 0)
				actual.Utime = 0
				for i := range actual.Models {
					require.True(t, actual.Models[i].Ctime > 0)
					actual.Models[i].Ctime = 0
					require.True(t, actual.Models[i].Utime > 0)
					actual.Models[i].Utime = 0
					require.True(t, actual.Models[i].Provider.Ctime > 0)
					actual.Models[i].Provider.Ctime = 0
					require.True(t, actual.Models[i].Provider.Utime > 0)
					actual.Models[i].Provider.Utime = 0
				}
				require.Equal(t, admin.ProviderVO{
					ID:     2010,
					Name:   "provider-2010",
					APIKey: "provider-apikey-2010",
					Models: []admin.ModelVO{
						{
							ID: 2002,
							Provider: admin.ProviderVO{
								ID:     2010,
								Name:   "provider-2010",
								APIKey: "provider-apikey-2010",
							},
							Name:        "model-2002",
							InputPrice:  2002,
							OutputPrice: 2002,
							PriceMode:   "2002",
						},
						{
							ID: 2001,
							Provider: admin.ProviderVO{
								ID:     2010,
								Name:   "provider-2010",
								APIKey: "provider-apikey-2010",
							},
							Name:        "model-2001",
							InputPrice:  2001,
							OutputPrice: 2001,
							PriceMode:   "2001",
						},
					},
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
			assertResultFunc: func(t *testing.T, r Result[admin.ProviderVO]) {
				t.Helper()
				require.Equal(t, Result[admin.ProviderVO]{Code: 501001, Msg: "系统错误"}, r)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost,
				"/providers/detail", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[admin.ProviderVO]()
			s.GinServer.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			tc.assertResultFunc(t, recorder.MustScan())
		})
	}
}

func (s *ProviderTestSuite) TestModel_Save() {
	t := s.T()

	testCases := []struct {
		name   string
		before func(t *testing.T)
		req    admin.ModelVO
		after  func(t *testing.T, req admin.ModelVO)

		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建成功",
			before: func(t *testing.T) {
				t.Helper()
				_, err := s.svc.SaveProvider(t.Context(), domain.Provider{
					ID:     2011,
					Name:   "provider-2011",
					APIKey: "provider-apikey-2011",
				})
				require.NoError(t, err)
			},
			req: admin.ModelVO{
				ID: 2011,
				Provider: admin.ProviderVO{
					ID:     2011,
					Name:   "provider-2011",
					APIKey: "provider-apikey-2011",
				},
				Name:        "model-2011",
				InputPrice:  2011,
				OutputPrice: 2011,
				PriceMode:   "2011",
			},
			after: func(t *testing.T, req admin.ModelVO) {
				t.Helper()
				p, err := s.svc.ModelDetail(t.Context(), 2011)
				assert.NoError(t, err)
				s.assertModel(t, req, p)
			},
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 2011,
			},
		},
		{
			name: "更新成功",
			before: func(t *testing.T) {
				t.Helper()

				pid, err := s.svc.SaveProvider(t.Context(), domain.Provider{
					ID:     2012,
					Name:   "provider-2012",
					APIKey: "provider-apikey-2012",
				})
				require.NoError(t, err)

				_, err = s.svc.SaveModel(t.Context(), domain.Model{
					ID:          2012,
					Provider:    domain.Provider{ID: pid},
					Name:        "model-2012",
					InputPrice:  2012,
					OutputPrice: 2012,
					PriceMode:   "2012",
				})
				require.NoError(t, err)

				_, err = s.svc.SaveProvider(t.Context(), domain.Provider{
					ID:     2013,
					Name:   "provider-2013",
					APIKey: "provider-apikey-2013",
				})
				require.NoError(t, err)
			},
			req: admin.ModelVO{
				ID: 2012,
				Provider: admin.ProviderVO{
					ID:     2013,
					Name:   "provider-2013",
					APIKey: "provider-apikey-2013",
				},
				Name:        "model-2013",
				InputPrice:  2013,
				OutputPrice: 2013,
				PriceMode:   "2013",
			},
			after: func(t *testing.T, req admin.ModelVO) {
				t.Helper()
				p, err := s.svc.ModelDetail(t.Context(), 2012)
				assert.NoError(t, err)
				s.assertModel(t, req, p)
			},
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 2012,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)

			req, err := http.NewRequest(http.MethodPost,
				"/models/save", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[int64]()
			s.GinServer.ServeHTTP(recorder, req)
			result := recorder.MustScan()
			assert.Equal(t, tc.wantCode, result.Code)
			require.Equal(t, tc.wantRes, result)

			tc.after(t, tc.req)
		})
	}
}

func (s *ProviderTestSuite) assertModel(t *testing.T, expected admin.ModelVO, actual domain.Model) {
	t.Helper()
	assert.Equal(t, expected.ID, actual.ID)
	s.assertProvider(t, expected.Provider, actual.Provider)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.InputPrice, actual.InputPrice)
	assert.Equal(t, expected.OutputPrice, actual.OutputPrice)
	assert.Equal(t, expected.PriceMode, actual.PriceMode)
	if expected.Ctime != 0 {
		assert.Equal(t, expected.Ctime, actual.Ctime)
	} else {
		assert.True(t, actual.Ctime > 0)
	}
	if expected.Utime != 0 {
		assert.Equal(t, expected.Utime, actual.Utime)
	} else {
		assert.True(t, actual.Utime > 0)
	}
}

func (s *ProviderTestSuite) TestModel_Detail() {
	t := s.T()

	testCases := []struct {
		name             string
		before           func(t *testing.T)
		req              admin.IDReq
		wantCode         int
		assertResultFunc func(t *testing.T, r Result[admin.ModelVO])
	}{
		{
			name: "id存在",
			before: func(t *testing.T) {
				t.Helper()
				pid, err := s.svc.SaveProvider(t.Context(), domain.Provider{
					ID:     2014,
					Name:   "provider-2014",
					APIKey: "provider-apikey-2014",
				})
				require.NoError(t, err)

				_, err = s.svc.SaveModel(t.Context(), domain.Model{
					ID:          2013,
					Provider:    domain.Provider{ID: pid},
					Name:        "model-2013",
					InputPrice:  2013,
					OutputPrice: 2013,
					PriceMode:   "2013",
				})
				require.NoError(t, err)
			},
			req: admin.IDReq{
				ID: 2013,
			},
			wantCode: 200,
			assertResultFunc: func(t *testing.T, r Result[admin.ModelVO]) {
				t.Helper()
				actual := r.Data
				require.True(t, actual.Ctime > 0)
				actual.Ctime = 0
				require.True(t, actual.Utime > 0)
				actual.Utime = 0

				require.True(t, actual.Provider.Ctime > 0)
				actual.Provider.Ctime = 0
				require.True(t, actual.Provider.Utime > 0)
				actual.Provider.Utime = 0

				require.Equal(t, admin.ModelVO{
					ID: 2013,
					Provider: admin.ProviderVO{
						ID:     2014,
						Name:   "provider-2014",
						APIKey: "provider-apikey-2014",
					},
					Name:        "model-2013",
					InputPrice:  2013,
					OutputPrice: 2013,
					PriceMode:   "2013",
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
			assertResultFunc: func(t *testing.T, r Result[admin.ModelVO]) {
				t.Helper()
				require.Equal(t, Result[admin.ModelVO]{Code: 501001, Msg: "系统错误"}, r)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost,
				"/models/detail", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[admin.ModelVO]()
			s.GinServer.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			tc.assertResultFunc(t, recorder.MustScan())
		})
	}
}
