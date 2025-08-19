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

type InvocationConfigTestSuite struct {
	suite.Suite
	*testioc.TestApp
	repo *repository.InvocationConfigRepo
}

func TestInvocationConfigTestSuite(t *testing.T) {
	suite.Run(t, new(InvocationConfigTestSuite))
}

func (s *InvocationConfigTestSuite) SetupSuite() {
	app := testioc.InitApp(testioc.TestOnly{})
	s.TestApp = app
	s.repo = repository.NewInvocationConfigRepo(dao.NewInvocationConfigDAO(s.DB))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bizRepo := repository.NewBizConfigRepository(dao.NewBizConfigDAO(s.DB))
	_, err := bizRepo.Save(ctx, domain.BizConfig{
		ID:        1,
		Name:      "biz1",
		OwnerID:   1,
		OwnerType: domain.OwnerTypeOrganization.String(),
		Config:    "",
	})
	s.NoError(err)

	_, err = bizRepo.Save(ctx, domain.BizConfig{
		ID:        2,
		Name:      "biz2",
		OwnerID:   2,
		OwnerType: domain.OwnerTypeUser.String(),
		Config:    "",
	})
	s.NoError(err)

	_, err = bizRepo.Save(ctx, domain.BizConfig{
		ID:        3,
		Name:      "biz3",
		OwnerID:   3,
		OwnerType: domain.OwnerTypeUser.String(),
		Config:    "",
	})
	s.NoError(err)

	_, err = bizRepo.Save(ctx, domain.BizConfig{
		ID:        4,
		Name:      "biz4",
		OwnerID:   4,
		OwnerType: domain.OwnerTypeUser.String(),
		Config:    "",
	})
	s.NoError(err)

	providerRepo := repository.NewProviderRepository(dao.NewProviderDAO(s.DB))
	provider3 := domain.Provider{
		ID:     3,
		Name:   "provider3",
		APIKey: "apikey3",
	}
	_, err = providerRepo.SaveProvider(ctx, provider3)
	s.NoError(err)
	provider4 := domain.Provider{
		ID:     4,
		Name:   "provider4",
		APIKey: "apikey4",
	}
	_, err = providerRepo.SaveProvider(ctx, provider4)
	s.NoError(err)
	_, err = providerRepo.SaveModel(ctx, domain.Model{
		ID:          1,
		Name:        "model1",
		InputPrice:  10,
		OutputPrice: 20,
		PriceMode:   "mode-1",
		Provider:    provider3,
	})
	s.NoError(err)

	_, err = providerRepo.SaveModel(ctx, domain.Model{
		ID:          2,
		Name:        "model2",
		InputPrice:  20,
		OutputPrice: 40,
		PriceMode:   "mode-2",
		Provider:    provider3,
	})
	s.NoError(err)

	_, err = providerRepo.SaveModel(ctx, domain.Model{
		ID:          3,
		Name:        "model3",
		InputPrice:  30,
		OutputPrice: 60,
		PriceMode:   "mode-3",
		Provider:    provider4,
	})
	s.NoError(err)
	_, err = providerRepo.SaveModel(ctx, domain.Model{
		ID:          4,
		Name:        "model4",
		InputPrice:  40,
		OutputPrice: 80,
		PriceMode:   "mode-4",
		Provider:    provider4,
	})
	s.NoError(err)
}

func (s *InvocationConfigTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE biz_configs").Error
	s.NoError(err)
	err = s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE providers").Error
	s.NoError(err)
	err = s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE models").Error
	s.NoError(err)
	err = s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE invocation_configs").Error
	s.NoError(err)
	err = s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE invocation_config_versions").Error
	s.NoError(err)
}

func (s *InvocationConfigTestSuite) TestConfig_Save() {
	t := s.T()

	testCases := []struct {
		name     string
		before   func(t *testing.T)
		req      admin.InvocationConfigVO
		after    func(t *testing.T, expected admin.InvocationConfigVO)
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name:   "新建成功",
			before: func(t *testing.T) {},
			req: admin.InvocationConfigVO{
				ID:          1,
				Name:        "test-invocation-1",
				BizID:       1,
				Description: "test-invocation-1",
			},
			after: func(t *testing.T, expected admin.InvocationConfigVO) {
				t.Helper()
				actual, err := s.repo.Get(t.Context(), 1)
				require.NoError(t, err)
				s.assertConfig(t, expected, actual)
			},
			wantCode: 200,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 1,
			},
		},
		{
			name: "更新成功",
			before: func(t *testing.T) {
				t.Helper()
				now := time.Now()
				_, err := s.repo.Save(t.Context(), domain.InvocationConfig{
					ID:   2,
					Name: "test-invocation-2",
					Biz: domain.BizConfig{
						ID: 1,
					},
					Description: "test-invocation-2",
					Ctime:       now,
					Utime:       now,
				})
				require.NoError(t, err)
			},
			req: admin.InvocationConfigVO{
				ID:          2,
				Name:        "update-test-invocation-2",
				BizID:       2,
				Description: "update-test-invocation-2",
			},
			after: func(t *testing.T, expected admin.InvocationConfigVO) {
				t.Helper()
				actual, err := s.repo.Get(t.Context(), 2)
				require.NoError(t, err)
				s.assertConfig(t, expected, actual)
			},
			wantCode: 200,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 2,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t, tc.req)

			req, err := http.NewRequest(http.MethodPost,
				"/invocation-configs/save", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[int64]()
			s.GinServer.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			require.Equal(t, tc.wantRes, recorder.MustScan())
		})
	}
}

func (s *InvocationConfigTestSuite) assertConfig(t *testing.T, expected admin.InvocationConfigVO, actual domain.InvocationConfig) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.BizID, actual.Biz.ID)
	assert.Equal(t, expected.BizName, actual.Biz.Name)
	assert.Equal(t, expected.Description, actual.Description)
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

func (s *InvocationConfigTestSuite) TestConfig_List() {
	t := s.T()

	err := s.DB.Exec("TRUNCATE TABLE `invocation_configs`").Error
	require.NoError(t, err)

	total := 6
	expected := make([]admin.InvocationConfigVO, total)
	for i := range total {
		id := int64(3 + i)
		_, err := s.repo.Save(t.Context(), domain.InvocationConfig{
			ID:   id,
			Name: fmt.Sprintf("test-invocation-%d", id),
			Biz: domain.BizConfig{
				ID: 1,
			},
			Description: fmt.Sprintf("test-invocation-%d", id),
		})
		require.NoError(t, err)
		expected[total-i-1] = admin.InvocationConfigVO{
			ID:          id,
			Name:        fmt.Sprintf("test-invocation-%d", id),
			BizID:       1,
			Description: fmt.Sprintf("test-invocation-%d", id),
		}
	}

	req := admin.ListReq{
		Offset: 0,
		Limit:  5,
	}

	httpReq, err := http.NewRequest(http.MethodPost,
		"/invocation-configs/list", iox.NewJSONReader(req))
	require.NoError(t, err)
	httpReq.Header.Set("content-type", "application/json")
	recorder := NewJSONResponseRecorder[ginx.DataList[admin.InvocationConfigVO]]()
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

func (s *InvocationConfigTestSuite) TestConfig_Detail() {
	t := s.T()

	testCases := []struct {
		name             string
		before           func(t *testing.T)
		req              admin.IDReq
		wantCode         int
		assertResultFunc func(t *testing.T, r Result[admin.InvocationConfigVO])
	}{
		{
			name: "id存在",
			before: func(t *testing.T) {
				t.Helper()
				_, err := s.repo.Save(t.Context(), domain.InvocationConfig{
					ID:   9,
					Name: "test-invocation-9",
					Biz: domain.BizConfig{
						ID:   2,
						Name: "biz2",
					},
					Description: "test-invocation-9",
				})
				require.NoError(t, err)
			},
			req: admin.IDReq{
				ID: 9,
			},
			wantCode: 200,
			assertResultFunc: func(t *testing.T, r Result[admin.InvocationConfigVO]) {
				t.Helper()
				actual := r.Data
				require.NotZero(t, actual.Utime)
				require.NotZero(t, actual.Ctime)
				actual.Utime, actual.Ctime = 0, 0
				require.Equal(t, admin.InvocationConfigVO{
					ID:          9,
					Name:        "test-invocation-9",
					BizID:       2,
					BizName:     "biz2",
					Description: "test-invocation-9",
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
			assertResultFunc: func(t *testing.T, r Result[admin.InvocationConfigVO]) {
				t.Helper()
				require.Equal(t, Result[admin.InvocationConfigVO]{Code: 501001, Msg: "系统错误"}, r)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost,
				"/invocation-configs/detail", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[admin.InvocationConfigVO]()
			s.GinServer.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			tc.assertResultFunc(t, recorder.MustScan())
		})
	}
}

func (s *InvocationConfigTestSuite) TestVersion_Save() {
	t := s.T()
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		req      admin.InvocationConfigVersionVO
		after    func(t *testing.T, expected admin.InvocationConfigVersionVO)
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建成功",
			before: func(t *testing.T) {
				t.Helper()
				_, err := s.repo.Save(t.Context(), domain.InvocationConfig{
					ID:   10,
					Name: "test-invocation-10",
					Biz: domain.BizConfig{
						ID:   3,
						Name: "biz3",
					},
					Description: "test-invocation-10",
				})
				require.NoError(t, err)
			},
			req: admin.InvocationConfigVersionVO{

				ID:           1,
				InvID:        10,
				ModelID:      1,
				Version:      "v1",
				Prompt:       "prompt1",
				SystemPrompt: "systemPrompt1",
				JSONSchema:   "jsonSchema1",
				Temperature:  1,
				TopP:         2,
				MaxTokens:    100,
				Status:       domain.InvocationCfgVersionStatusDraft.String(),
			},
			after: func(t *testing.T, expected admin.InvocationConfigVersionVO) {
				t.Helper()
				actual, err := s.repo.GetVersionByID(t.Context(), 1)
				require.NoError(t, err)
				s.assertVersion(t, expected, actual)
			},
			wantCode: 200,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 1,
			},
		},
		{
			name: "更新成功",
			before: func(t *testing.T) {
				t.Helper()
				_, err := s.repo.Save(t.Context(), domain.InvocationConfig{
					ID:   11,
					Name: "test-invocation-11",
					Biz: domain.BizConfig{
						ID:   3,
						Name: "biz3",
					},
					Description: "test-invocation-11",
				})
				require.NoError(t, err)

				_, err = s.repo.SaveVersion(t.Context(), domain.InvocationConfigVersion{
					ID: 2,
					Config: domain.InvocationConfig{
						ID: 11,
					},
					Model: domain.Model{
						ID: 1,
					},
					Version:      "v2",
					Prompt:       "prompt2",
					SystemPrompt: "systemPrompt2",
					JSONSchema:   "jsonSchema2",
					Temperature:  2,
					TopP:         2,
					MaxTokens:    200,
					Status:       domain.InvocationCfgVersionStatusActive,
				})
				require.NoError(t, err)
			},
			req: admin.InvocationConfigVersionVO{
				ID:           2,
				InvID:        11,
				ModelID:      3,
				Version:      "v3",
				Prompt:       "prompt3",
				SystemPrompt: "systemPrompt3",
				JSONSchema:   "jsonSchema3",
				Temperature:  3,
				TopP:         3,
				MaxTokens:    300,
				Status:       domain.InvocationCfgVersionStatusDraft.String(),
			},
			after: func(t *testing.T, expected admin.InvocationConfigVersionVO) {
				t.Helper()
				actual, err := s.repo.GetVersionByID(t.Context(), 2)
				require.NoError(t, err)
				s.assertVersion(t, expected, actual)
			},
			wantCode: 200,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 2,
			},
		},
		{
			name:   "状态非法",
			before: func(t *testing.T) {},
			req: admin.InvocationConfigVersionVO{
				ID:                1,
				InvID:             10,
				ModelID:           1,
				ModelName:         "mode1",
				ModelProviderID:   3,
				ModelProviderName: "provider1",
				Version:           "v1",
				Prompt:            "prompt1",
				SystemPrompt:      "systemPrompt1",
				JSONSchema:        "jsonSchema1",
				Temperature:       1,
				TopP:              2,
				MaxTokens:         100,
				Status:            "invalid",
			},
			after:    func(t *testing.T, expected admin.InvocationConfigVersionVO) {},
			wantCode: 500,
			wantRes: Result[int64]{
				Code: 501001, Msg: "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t, tc.req)
			req, err := http.NewRequest(http.MethodPost,
				"/invocation-configs/versions/save", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[int64]()
			s.GinServer.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantRes, recorder.MustScan())
		})
	}
}

func (s *InvocationConfigTestSuite) assertVersion(t *testing.T, expected admin.InvocationConfigVersionVO, actual domain.InvocationConfigVersion) {
	t.Helper()
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.InvID, actual.Config.ID)
	assert.Equal(t, expected.ModelID, actual.Model.ID)
	assert.Equal(t, expected.ModelName, actual.Model.Name)
	assert.Equal(t, expected.ModelProviderID, actual.Model.Provider.ID)
	assert.Equal(t, expected.ModelProviderName, actual.Model.Provider.Name)
	assert.Equal(t, expected.Version, actual.Version)
	assert.Equal(t, expected.Prompt, actual.Prompt)
	assert.Equal(t, expected.SystemPrompt, actual.SystemPrompt)
	assert.Equal(t, expected.JSONSchema, actual.JSONSchema)
	assert.Equal(t, expected.Temperature, actual.Temperature)
	assert.Equal(t, expected.TopP, actual.TopP)
	assert.Equal(t, expected.MaxTokens, actual.MaxTokens)
	assert.Equal(t, expected.Status, actual.Status.String())
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

func (s *InvocationConfigTestSuite) TestVersion_List() {
	t := s.T()

	err := s.DB.Exec("TRUNCATE TABLE `invocation_config_versions`").Error
	require.NoError(t, err)

	invID, err := s.repo.Save(t.Context(), domain.InvocationConfig{
		ID:          12,
		Name:        "test-invocation-12",
		Biz:         domain.BizConfig{ID: 1},
		Description: "test-invocation-12",
	})
	require.NoError(t, err)

	total := 6
	expected := make([]admin.InvocationConfigVersionVO, total)
	for i := range total {
		id := int64(3 + i)
		_, err := s.repo.SaveVersion(t.Context(), domain.InvocationConfigVersion{
			ID:     id,
			Config: domain.InvocationConfig{ID: invID},
			Model: domain.Model{
				ID: 1,
			},
			Version:      fmt.Sprintf("v%d", id),
			Prompt:       fmt.Sprintf("prompt%d", id),
			SystemPrompt: fmt.Sprintf("systemPrompt%d", id),
			JSONSchema:   fmt.Sprintf("jsonSchema%d", id),
			Temperature:  float32(id),
			TopP:         float32(id),
			MaxTokens:    int(id * 100),
			Status:       domain.InvocationCfgVersionStatusDraft,
		})
		require.NoError(t, err)
		expected[total-i-1] = admin.InvocationConfigVersionVO{
			ID:           id,
			InvID:        invID,
			ModelID:      1,
			Version:      fmt.Sprintf("v%d", id),
			Prompt:       fmt.Sprintf("prompt%d", id),
			SystemPrompt: fmt.Sprintf("systemPrompt%d", id),
			JSONSchema:   fmt.Sprintf("jsonSchema%d", id),
			Temperature:  float32(id),
			TopP:         float32(id),
			MaxTokens:    int(id * 100),
			Status:       domain.InvocationCfgVersionStatusDraft.String(),
		}
	}

	req := admin.ListInvocationConfigVersionsReq{
		InvID:  invID,
		Offset: 0,
		Limit:  5,
	}

	httpReq, err := http.NewRequest(http.MethodPost,
		"/invocation-configs/versions/list", iox.NewJSONReader(req))
	require.NoError(t, err)
	httpReq.Header.Set("content-type", "application/json")
	recorder := NewJSONResponseRecorder[ginx.DataList[admin.InvocationConfigVersionVO]]()
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

func (s *InvocationConfigTestSuite) TestVersion_Detail() {
	t := s.T()

	testCases := []struct {
		name             string
		before           func(t *testing.T)
		req              admin.IDReq
		wantCode         int
		assertResultFunc func(t *testing.T, r Result[admin.InvocationConfigVersionVO])
	}{
		{
			name: "id存在",
			before: func(t *testing.T) {
				t.Helper()
				_, err := s.repo.Save(t.Context(), domain.InvocationConfig{
					ID:   13,
					Name: "test-invocation-13",
					Biz: domain.BizConfig{
						ID:   2,
						Name: "biz2",
					},
					Description: "test-invocation-13",
				})
				require.NoError(t, err)

				_, err = s.repo.SaveVersion(t.Context(), domain.InvocationConfigVersion{
					ID:     9,
					Config: domain.InvocationConfig{ID: 13},
					Model: domain.Model{
						ID: 2,
					},
					Version:      fmt.Sprintf("v%d", 9),
					Prompt:       fmt.Sprintf("prompt%d", 9),
					SystemPrompt: fmt.Sprintf("systemPrompt%d", 9),
					JSONSchema:   fmt.Sprintf("jsonSchema%d", 9),
					Temperature:  float32(9),
					TopP:         float32(9),
					MaxTokens:    9 * 100,
					Status:       domain.InvocationCfgVersionStatusActive,
				})
				require.NoError(t, err)
			},
			req: admin.IDReq{
				ID: 9,
			},
			wantCode: 200,
			assertResultFunc: func(t *testing.T, r Result[admin.InvocationConfigVersionVO]) {
				t.Helper()
				actual := r.Data
				require.NotZero(t, actual.Utime)
				require.NotZero(t, actual.Ctime)
				actual.Utime, actual.Ctime = 0, 0
				require.Equal(t, admin.InvocationConfigVersionVO{
					ID:                9,
					InvID:             13,
					ModelID:           2,
					ModelProviderID:   3,
					ModelProviderName: "provider3",
					ModelName:         "model2",
					Version:           fmt.Sprintf("v%d", 9),
					Prompt:            fmt.Sprintf("prompt%d", 9),
					SystemPrompt:      fmt.Sprintf("systemPrompt%d", 9),
					JSONSchema:        fmt.Sprintf("jsonSchema%d", 9),
					Temperature:       float32(9),
					TopP:              float32(9),
					MaxTokens:         9 * 100,
					Status:            domain.InvocationCfgVersionStatusActive.String(),
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
			assertResultFunc: func(t *testing.T, r Result[admin.InvocationConfigVersionVO]) {
				t.Helper()
				require.Equal(t, Result[admin.InvocationConfigVersionVO]{Code: 501001, Msg: "系统错误"}, r)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost,
				"/invocation-configs/versions/detail", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[admin.InvocationConfigVersionVO]()
			s.GinServer.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			tc.assertResultFunc(t, recorder.MustScan())
		})
	}
}

func (s *InvocationConfigTestSuite) TestVersion_Activate() {
	t := s.T()

	testCases := []struct {
		name     string
		before   func(t *testing.T)
		req      admin.IDReq
		after    func(t *testing.T)
		wantCode int
		wantRes  Result[any]
	}{
		{
			name: "不存在处于活跃状态的版本",
			before: func(t *testing.T) {
				t.Helper()
				_, err := s.repo.Save(t.Context(), domain.InvocationConfig{
					ID:   14,
					Name: "test-invocation-14",
					Biz: domain.BizConfig{
						ID:   2,
						Name: "biz2",
					},
					Description: "test-invocation-14",
				})
				require.NoError(t, err)

				_, err = s.repo.SaveVersion(t.Context(), domain.InvocationConfigVersion{
					ID:     10,
					Config: domain.InvocationConfig{ID: 14},
					Model: domain.Model{
						ID: 2,
					},
					Version:      fmt.Sprintf("v%d", 10),
					Prompt:       fmt.Sprintf("prompt%d", 10),
					SystemPrompt: fmt.Sprintf("systemPrompt%d", 10),
					JSONSchema:   fmt.Sprintf("jsonSchema%d", 10),
					Temperature:  float32(10),
					TopP:         float32(10),
					MaxTokens:    10 * 100,
					Status:       domain.InvocationCfgVersionStatusDraft,
				})
				require.NoError(t, err)

				_, err = s.repo.SaveVersion(t.Context(), domain.InvocationConfigVersion{
					ID:     11,
					Config: domain.InvocationConfig{ID: 14},
					Model: domain.Model{
						ID: 2,
					},
					Version:      fmt.Sprintf("v%d", 11),
					Prompt:       fmt.Sprintf("prompt%d", 11),
					SystemPrompt: fmt.Sprintf("systemPrompt%d", 11),
					JSONSchema:   fmt.Sprintf("jsonSchema%d", 11),
					Temperature:  float32(11),
					TopP:         float32(11),
					MaxTokens:    12 * 100,
					Status:       domain.InvocationCfgVersionStatusDraft,
				})
				require.NoError(t, err)
			},
			req: admin.IDReq{
				ID: 10,
			},
			after: func(t *testing.T) {
				t.Helper()
				versions, err := s.repo.ListVersions(t.Context(), 14, 0, 3)
				require.NoError(t, err)
				for i := range versions {
					if versions[i].ID == 10 {
						require.Equal(t, domain.InvocationCfgVersionStatusActive, versions[i].Status)
					} else {
						require.Equal(t, domain.InvocationCfgVersionStatusDraft, versions[i].Status)
					}
				}
			},
			wantCode: 200,
			wantRes: Result[any]{
				Msg: "OK",
			},
		},
		{
			name: "已存在处于活跃状态的版本，变更为其他版本",
			before: func(t *testing.T) {
				t.Helper()
				_, err := s.repo.Save(t.Context(), domain.InvocationConfig{
					ID:   15,
					Name: "test-invocation-15",
					Biz: domain.BizConfig{
						ID:   2,
						Name: "biz2",
					},
					Description: "test-invocation-15",
				})
				require.NoError(t, err)

				_, err = s.repo.SaveVersion(t.Context(), domain.InvocationConfigVersion{
					ID:     12,
					Config: domain.InvocationConfig{ID: 15},
					Model: domain.Model{
						ID: 2,
					},
					Version:      fmt.Sprintf("v%d", 12),
					Prompt:       fmt.Sprintf("prompt%d", 12),
					SystemPrompt: fmt.Sprintf("systemPrompt%d", 12),
					JSONSchema:   fmt.Sprintf("jsonSchema%d", 12),
					Temperature:  float32(12),
					TopP:         float32(12),
					MaxTokens:    12 * 100,
					Status:       domain.InvocationCfgVersionStatusDraft,
				})
				require.NoError(t, err)

				_, err = s.repo.SaveVersion(t.Context(), domain.InvocationConfigVersion{
					ID:     13,
					Config: domain.InvocationConfig{ID: 15},
					Model: domain.Model{
						ID: 2,
					},
					Version:      fmt.Sprintf("v%d", 13),
					Prompt:       fmt.Sprintf("prompt%d", 13),
					SystemPrompt: fmt.Sprintf("systemPrompt%d", 13),
					JSONSchema:   fmt.Sprintf("jsonSchema%d", 13),
					Temperature:  float32(13),
					TopP:         float32(13),
					MaxTokens:    13 * 100,
					Status:       domain.InvocationCfgVersionStatusActive,
				})
				require.NoError(t, err)

				_, err = s.repo.SaveVersion(t.Context(), domain.InvocationConfigVersion{
					ID:     14,
					Config: domain.InvocationConfig{ID: 15},
					Model: domain.Model{
						ID: 2,
					},
					Version:      fmt.Sprintf("v%d", 14),
					Prompt:       fmt.Sprintf("prompt%d", 14),
					SystemPrompt: fmt.Sprintf("systemPrompt%d", 14),
					JSONSchema:   fmt.Sprintf("jsonSchema%d", 14),
					Temperature:  float32(14),
					TopP:         float32(14),
					MaxTokens:    14 * 100,
					Status:       domain.InvocationCfgVersionStatusDraft,
				})
				require.NoError(t, err)
			},
			req: admin.IDReq{
				ID: 12,
			},
			after: func(t *testing.T) {
				t.Helper()
				versions, err := s.repo.ListVersions(t.Context(), 15, 0, 3)
				require.NoError(t, err)
				for i := range versions {
					if versions[i].ID == 12 {
						require.Equal(t, domain.InvocationCfgVersionStatusActive, versions[i].Status)
					} else {
						require.Equal(t, domain.InvocationCfgVersionStatusDraft, versions[i].Status)
					}
				}
			},
			wantCode: 200,
			wantRes: Result[any]{
				Msg: "OK",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			req, err := http.NewRequest(http.MethodPost,
				"/invocation-configs/versions/activate", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[any]()
			s.GinServer.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			require.Equal(t, tc.wantRes, recorder.MustScan())
		})
	}
}

func (s *InvocationConfigTestSuite) TestVersion_Fork() {
	t := s.T()

	testCases := []struct {
		name         string
		before       func(t *testing.T)
		req          admin.IDReq
		after        func(t *testing.T, id int64)
		wantCode     int
		assertResult func(t *testing.T, r Result[int64])
	}{
		{
			name: "成功",
			before: func(t *testing.T) {
				t.Helper()
				_, err := s.repo.Save(t.Context(), domain.InvocationConfig{
					ID:   16,
					Name: "test-invocation-16",
					Biz: domain.BizConfig{
						ID:   2,
						Name: "biz2",
					},
					Description: "test-invocation-16",
				})
				require.NoError(t, err)

				_, err = s.repo.SaveVersion(t.Context(), domain.InvocationConfigVersion{
					ID:     15,
					Config: domain.InvocationConfig{ID: 16},
					Model: domain.Model{
						ID: 2,
					},
					Version:      fmt.Sprintf("v%d", 15),
					Prompt:       fmt.Sprintf("prompt%d", 15),
					SystemPrompt: fmt.Sprintf("systemPrompt%d", 15),
					JSONSchema:   fmt.Sprintf("jsonSchema%d", 15),
					Temperature:  float32(15),
					TopP:         float32(15),
					MaxTokens:    15 * 100,
					Status:       domain.InvocationCfgVersionStatusActive,
				})
				require.NoError(t, err)
			},
			req: admin.IDReq{
				ID: 15,
			},
			after: func(t *testing.T, id int64) {
				t.Helper()
				version, err := s.repo.GetVersionByID(t.Context(), id)
				require.NoError(t, err)
				require.NotZero(t, version.Ctime)
				require.NotZero(t, version.Utime)
				now := time.Now()
				version.Ctime, version.Utime = now, now
				require.Equal(t, domain.InvocationConfigVersion{
					ID:     id,
					Config: domain.InvocationConfig{ID: 16},
					Model: domain.Model{
						ID: 2,
					},
					Version:      fmt.Sprintf("v%d", 15),
					Prompt:       fmt.Sprintf("prompt%d", 15),
					SystemPrompt: fmt.Sprintf("systemPrompt%d", 15),
					JSONSchema:   fmt.Sprintf("jsonSchema%d", 15),
					Temperature:  float32(15),
					TopP:         float32(15),
					MaxTokens:    15 * 100,
					Status:       domain.InvocationCfgVersionStatusDraft,
					Ctime:        now,
					Utime:        now,
				}, version)
			},
			wantCode: 200,
			assertResult: func(t *testing.T, r Result[int64]) {
				require.Equal(t, "OK", r.Msg)
				require.Greater(t, r.Data, int64(15))
			},
		},
		{
			name:   "失败_要Fork的版本ID不存在",
			before: func(t *testing.T) {},
			req: admin.IDReq{
				ID: 10000,
			},
			after:    func(t *testing.T, _ int64) {},
			wantCode: 500,
			assertResult: func(t *testing.T, r Result[int64]) {
				require.Equal(t, Result[int64]{Code: 501001, Msg: "系统错误"}, r)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost,
				"/invocation-configs/versions/fork", iox.NewJSONReader(tc.req))
			require.NoError(t, err)
			req.Header.Set("content-type", "application/json")
			recorder := NewJSONResponseRecorder[int64]()
			s.GinServer.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			result := recorder.MustScan()
			tc.assertResult(t, result)
			tc.after(t, result.Data)
		})
	}
}
