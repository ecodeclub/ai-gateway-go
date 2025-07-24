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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/admin"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/test/mocks"
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
	handler := admin.NewQuotaHandler(svc)
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
			// 每个测试用例开始时清理数据
			err := q.db.Exec("TRUNCATE TABLE quotas").Error
			require.NoError(t, err)
			err = q.db.Exec("TRUNCATE TABLE quota_records").Error
			require.NoError(t, err)
			err = q.db.Exec("TRUNCATE TABLE temp_quotas").Error
			require.NoError(t, err)

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
		before  func()
		after   func()
		reqBody string
	}{
		{
			name: "创建临时额度",
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
				err := q.db.Where("uid = ? AND `key` = ?", 1, "temp_key_1").First(&quota).Error
				require.NoError(t, err)
				assert.Equal(t, int64(100000), quota.Amount)
				assert.Equal(t, int64(123), quota.StartTime)
				assert.Equal(t, int64(456), quota.EndTime)
			},
			reqBody: `{"amount": 100000, "key": "temp_key_1", "start_time": 123, "end_time": 456}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// 每个测试用例开始时清理数据
			err := q.db.Exec("TRUNCATE TABLE quotas").Error
			require.NoError(t, err)
			err = q.db.Exec("TRUNCATE TABLE quota_records").Error
			require.NoError(t, err)
			err = q.db.Exec("TRUNCATE TABLE temp_quotas").Error
			require.NoError(t, err)

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
		before  func()
		after   func()
		reqBody string
	}{
		{
			name: "从主额度扣减",
			before: func() {
				sess := mocks.NewMockSession(ctrl)
				sess.EXPECT().Claims().Return(session.Claims{
					Uid: 1,
				}).AnyTimes()
				provider := mocks.NewMockProvider(ctrl)
				session.SetDefaultProvider(provider)
				provider.EXPECT().Get(gomock.Any()).Return(sess, nil)

				// 创建主额度
				quota := dao.Quota{Amount: 100, Key: "main_quota", UID: 1}
				err := q.db.Create(&quota).Error
				require.NoError(t, err)
			},
			after: func() {
				// 验证主额度被扣减
				var quota dao.Quota
				err := q.db.Where("uid = ? AND `key` = ?", 1, "main_quota").First(&quota).Error
				require.NoError(t, err)
				assert.Equal(t, int64(80), quota.Amount)

				// 验证扣减记录被创建
				var record dao.QuotaRecord
				err = q.db.Where("uid = ? AND `key` = ?", 1, "deduct_key_1").First(&record).Error
				require.NoError(t, err)
				assert.Equal(t, int64(20), record.Amount)
			},
			reqBody: `{"amount": 20, "key": "deduct_key_1"}`,
		},
		{
			name: "从临时额度扣减",
			before: func() {
				sess := mocks.NewMockSession(ctrl)
				sess.EXPECT().Claims().Return(session.Claims{
					Uid: 1,
				}).AnyTimes()
				provider := mocks.NewMockProvider(ctrl)
				session.SetDefaultProvider(provider)
				provider.EXPECT().Get(gomock.Any()).Return(sess, nil)

				// 创建临时额度
				now := time.Now().Unix()
				tempQuota := dao.TempQuota{
					Amount:    50,
					Key:       "temp_quota_1",
					UID:       1,
					StartTime: now,
					EndTime:   now + 24*3600,
				}
				err := q.db.Create(&tempQuota).Error
				require.NoError(t, err)
			},
			after: func() {
				// 验证临时额度被扣减
				var tempQuota dao.TempQuota
				err := q.db.Where("uid = ? AND `key` = ?", 1, "temp_quota_1").First(&tempQuota).Error
				require.NoError(t, err)
				assert.Equal(t, int64(30), tempQuota.Amount)

				// 验证扣减记录被创建
				var record dao.QuotaRecord
				err = q.db.Where("uid = ? AND `key` = ?", 1, "deduct_key_2").First(&record).Error
				require.NoError(t, err)
				assert.Equal(t, int64(20), record.Amount)
			},
			reqBody: `{"amount": 20, "key": "deduct_key_2"}`,
		},
		{
			name: "优先从临时额度扣减，不足再从主额度扣减",
			before: func() {
				sess := mocks.NewMockSession(ctrl)
				sess.EXPECT().Claims().Return(session.Claims{
					Uid: 1,
				}).AnyTimes()
				provider := mocks.NewMockProvider(ctrl)
				session.SetDefaultProvider(provider)
				provider.EXPECT().Get(gomock.Any()).Return(sess, nil)

				// 创建主额度
				quota := dao.Quota{Amount: 100, Key: "main_quota_2", UID: 1}
				err := q.db.Create(&quota).Error
				require.NoError(t, err)

				// 创建临时额度（金额不足）
				now := time.Now().Unix()
				tempQuota := dao.TempQuota{
					Amount:    10,
					Key:       "temp_quota_2",
					UID:       1,
					StartTime: now,
					EndTime:   now + 24*3600,
				}
				err = q.db.Create(&tempQuota).Error
				require.NoError(t, err)
			},
			after: func() {
				// 验证临时额度被完全扣减
				var tempQuota dao.TempQuota
				err := q.db.Where("uid = ? AND `key` = ?", 1, "temp_quota_2").First(&tempQuota).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), tempQuota.Amount)

				// 验证主额度被扣减
				var quota dao.Quota
				err = q.db.Where("uid = ? AND `key` = ?", 1, "main_quota_2").First(&quota).Error
				require.NoError(t, err)
				assert.Equal(t, int64(90), quota.Amount)

				// 验证扣减记录被创建
				var record dao.QuotaRecord
				err = q.db.Where("uid = ? AND `key` = ?", 1, "deduct_key_3").First(&record).Error
				require.NoError(t, err)
				assert.Equal(t, int64(30), record.Amount)
			},
			reqBody: `{"amount": 30, "key": "deduct_key_3"}`,
		},
		{
			name: "扣减失败 - 余额不足",
			before: func() {
				sess := mocks.NewMockSession(ctrl)
				sess.EXPECT().Claims().Return(session.Claims{
					Uid: 1,
				}).AnyTimes()
				provider := mocks.NewMockProvider(ctrl)
				session.SetDefaultProvider(provider)
				provider.EXPECT().Get(gomock.Any()).Return(sess, nil)

				// 创建少量主额度
				quota := dao.Quota{Amount: 10, Key: "main_quota_3", UID: 1}
				err := q.db.Create(&quota).Error
				require.NoError(t, err)
			},
			after: func() {
				// 验证主额度没有被扣减
				var quota dao.Quota
				err := q.db.Where("uid = ? AND `key` = ?", 1, "main_quota_3").First(&quota).Error
				require.NoError(t, err)
				assert.Equal(t, int64(10), quota.Amount) // 额度应该保持不变

				// 验证扣减记录没有被创建（因为事务回滚）
				var record dao.QuotaRecord
				err = q.db.Where("uid = ? AND `key` = ?", 1, "deduct_key_4").First(&record).Error
				assert.Error(t, err) // 应该找不到记录，因为事务回滚了
			},
			reqBody: `{"amount": 50, "key": "deduct_key_4"}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// 每个测试用例开始时清理数据
			err := q.db.Exec("TRUNCATE TABLE quotas").Error
			require.NoError(t, err)
			err = q.db.Exec("TRUNCATE TABLE quota_records").Error
			require.NoError(t, err)
			err = q.db.Exec("TRUNCATE TABLE temp_quotas").Error
			require.NoError(t, err)

			tc.before()
			req, err := http.NewRequest(http.MethodPost, "/deduct", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			q.server.ServeHTTP(resp, req)

			if tc.name == "扣减失败 - 余额不足" {
				assert.Equal(t, http.StatusOK, resp.Code) // HTTP状态码仍然是200

				var response map[string]interface{}
				err = json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, float64(400002), response["code"])
				assert.Equal(t, "余额不足", response["msg"])
			} else {
				assert.Equal(t, http.StatusOK, resp.Code)
			}

			tc.after()
		})
	}
}
