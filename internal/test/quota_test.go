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

package test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/test/mocks"
	"github.com/ecodeclub/ai-gateway-go/internal/web"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yumosx/got/pkg/config"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

type QuotaSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

func TestQuota(t *testing.T) {
	suite.Run(t, &QuotaSuite{})
}

func (q *QuotaSuite) SetupSuite() {
	dbConfig := config.NewConfig(
		config.WithDBName("ai_gateway_platform"),
		config.WithUserName("root"),
		config.WithPassword("root"),
		config.WithHost("127.0.0.1"),
		config.WithPort("13306"),
	)
	db, err := config.NewDB(dbConfig)
	require.NoError(q.T(), err)
	err = dao.InitQuotaTable(db)
	require.NoError(q.T(), err)
	q.db = db

	d := dao.NewQuotaDao(db)
	repo := repository.NewQuotaRepo(d)
	svc := service.NewQuotaService(repo)
	handler := web.NewQuotaHandler(svc)
	server := gin.Default()
	handler.PrivateRoutes(server)
	q.server = server
}

func (q *QuotaSuite) TearDownTest() {
	err := q.db.Exec("TRUNCATE TABLE quotas").Error
	require.NoError(q.T(), err)
	err = q.db.Exec("TRUNCATE TABLE quota_records").Error
	require.NoError(q.T(), err)
	err = q.db.Exec("TRUNCATE TABLE temp_quotas").Error
	require.NoError(q.T(), err)
}

func (q *QuotaSuite) TestQuotaSave() {
	t := q.T()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testcases := []struct {
		name    string
		before  func()
		after   func()
		reqBody string
	}{
		{
			name: "创建一个 quota",
			before: func() {
				sess := mocks.NewMockSession(ctrl)
				sess.EXPECT().Claims().Return(session.Claims{
					Uid: 1,
				}).AnyTimes()
				provider := mocks.NewMockProvider(ctrl)
				session.SetDefaultProvider(provider)
				provider.EXPECT().Get(gomock.Any()).Return(sess, nil)
			},
			after: func() {
				var quota dao.Quota
				err := q.db.Where("id = ?", 1).First(&quota).Error
				require.NoError(t, err)
				assert.Equal(t, int64(100000), quota.Amount)

				var record dao.QuotaRecord
				err = q.db.Where("id = ?", 1).First(&record).Error
				require.NoError(t, err)
				assert.Equal(t, "23911", record.Key)
			},
			reqBody: `{"amount": 100000, "key": "23911"}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodPost, "/quota/save", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			q.server.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)

			tc.after()
		})
	}
}

func (q *QuotaSuite) TestSaveTempQuota() {
	t := q.T()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testcases := []struct {
		name    string
		reqBody string
		before  func()
		after   func()
	}{
		{
			name:    "save temp",
			reqBody: `{"amount": 100000, "key": "23911", "start_time": "123", "end_time": "456"}`,
			before: func() {
				sess := mocks.NewMockSession(ctrl)
				sess.EXPECT().Claims().Return(session.Claims{
					Uid: 1,
				}).AnyTimes()
				provider := mocks.NewMockProvider(ctrl)
				session.SetDefaultProvider(provider)
				provider.EXPECT().Get(gomock.Any()).Return(sess, nil)
			},
			after: func() {
				var quota dao.TempQuota
				err := q.db.Where("id = ?", 1).First(&quota).Error
				require.NoError(t, err)
				assert.Equal(t, int64(100000), quota.Amount)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodPost, "/tmp/save", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			q.server.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)

			tc.after()
		})
	}
}

func (q *QuotaSuite) TestDeduct() {
	t := q.T()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testcases := []struct {
		name    string
		reqBody string
		before  func(db *gorm.DB, server *gin.Engine)
		after   func(db *gorm.DB)
	}{
		{
			name:    "deduct quota",
			reqBody: `{"amount": 10, "key": "23911"}`,
			before: func(db *gorm.DB, server *gin.Engine) {
				sess := mocks.NewMockSession(ctrl)
				sess.EXPECT().Claims().Return(session.Claims{
					Uid: 1,
				}).AnyTimes()
				provider := mocks.NewMockProvider(ctrl)
				session.SetDefaultProvider(provider)
				provider.EXPECT().Get(gomock.Any()).Return(sess, nil)

				quota := dao.Quota{Amount: 20, Key: "23911", UID: 1}
				db.Create(&quota)
			},
			after: func(db *gorm.DB) {
				var quota dao.Quota
				err := db.Where("id = ?", 1).First(&quota).Error
				require.NoError(t, err)
				assert.Equal(t, int64(10), quota.Amount)
			},
		},
		{
			name:    "deduct temp quota",
			reqBody: `{"amount": 10, "key": "23922"}`,
			before: func(db *gorm.DB, server *gin.Engine) {
				sess := mocks.NewMockSession(ctrl)
				sess.EXPECT().Claims().Return(session.Claims{
					Uid: 1,
				}).AnyTimes()
				provider := mocks.NewMockProvider(ctrl)
				session.SetDefaultProvider(provider)
				provider.EXPECT().Get(gomock.Any()).Return(sess, nil)

				quota := dao.TempQuota{Amount: 20, Key: "23922", UID: 1, StartTime: time.Now().Unix(), EndTime: time.Now().Add(24 * time.Hour).Unix()}
				err := db.Create(&quota).Error
				require.NoError(t, err)
			},
			after: func(db *gorm.DB) {
				var quota dao.TempQuota
				err := db.Where("id = ?", 1).First(&quota).Error
				require.NoError(t, err)
				assert.Equal(t, int64(10), quota.Amount)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// 独立初始化 DB 和 Gin Engine
			dbConfig := config.NewConfig(
				config.WithDBName("ai_gateway_platform"),
				config.WithUserName("root"),
				config.WithPassword("root"),
				config.WithHost("127.0.0.1"),
				config.WithPort("13306"),
			)
			db, err := config.NewDB(dbConfig)
			require.NoError(t, err)
			err = dao.InitQuotaTable(db)
			require.NoError(t, err)

			d := dao.NewQuotaDao(db)
			repo := repository.NewQuotaRepo(d)
			svc := service.NewQuotaService(repo)
			handler := web.NewQuotaHandler(svc)
			server := gin.Default()
			handler.PrivateRoutes(server)

			// 清理表
			defer func() {
				db.Exec("TRUNCATE TABLE quotas")
				db.Exec("TRUNCATE TABLE quota_records")
				db.Exec("TRUNCATE TABLE temp_quotas")
			}()

			tc.before(db, server)
			req, err := http.NewRequest(http.MethodPost, "/deduct", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			assert.Equal(t, http.StatusOK, resp.Code)
			tc.after(db)
		})
	}
}
