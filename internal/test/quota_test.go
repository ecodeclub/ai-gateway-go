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
	"testing"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	testioc "github.com/ecodeclub/ai-gateway-go/internal/test/ioc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type QuotaTestSuite struct {
	suite.Suite
	*testioc.TestApp
	dao *dao.QuotaDao
}

func TestQuota(t *testing.T) {
	suite.Run(t, &QuotaTestSuite{})
}

func (s *QuotaTestSuite) SetupSuite() {
	app := testioc.InitApp(testioc.TestOnly{})
	s.TestApp = app
	s.dao = dao.NewQuotaDao(s.DB)
}

func (s *QuotaTestSuite) TearDownTest() {
	t := s.T()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE temp_quotas").Error
	require.NoError(t, err)
	err = s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE quotas").Error
	require.NoError(t, err)
	err = s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE quota_records").Error
	require.NoError(t, err)
}

func (s *QuotaTestSuite) TestDeduct() {
	t := s.T()

	testcases := []struct {
		name   string
		uid    int64
		amount int64
		key    string
		before func(ctx context.Context, t *testing.T)
		after  func(ctx context.Context, t *testing.T)
	}{
		{
			name:   "扣减临时会员表",
			uid:    1,
			amount: 5,
			key:    "deduct-key-1",
			before: func(ctx context.Context, t *testing.T) {
				err := s.dao.CreateTempQuota(ctx, dao.TempQuota{
					StartTime: time.Now().Unix(),
					EndTime:   time.Now().Add(time.Hour * 24).Unix(),
					Key:       "temp-key-1",
					Amount:    10,
					UID:       1,
				})
				require.NoError(t, err)
			},
			after: func(ctx context.Context, t *testing.T) {
				list, err := s.dao.GetTempQuotaByUidAndTime(ctx, 1)
				require.NoError(t, err)
				require.Len(t, list, 1)
				assert.Equal(t, list[0].Amount, int64(5))
			},
		},
		{
			name:   "扣减永久会员表",
			uid:    2,
			amount: 5,
			key:    "deduct-key-2",
			before: func(ctx context.Context, t *testing.T) {
				err := s.dao.AddQuota(ctx, "quota-key-2", dao.Quota{
					UID:    2,
					Amount: 10,
				})
				require.NoError(t, err)
			},
			after: func(ctx context.Context, t *testing.T) {
				quota, err := s.dao.GetQuotaByUid(ctx, 2)
				require.NoError(t, err)
				assert.Equal(t, int64(5), quota.Amount)
			},
		},
		{
			name:   "先扣减临时会员表, 然后扣减永久会员表",
			uid:    3,
			amount: 13,
			key:    "deduct-key-3",
			before: func(ctx context.Context, t *testing.T) {
				err := s.dao.CreateTempQuota(ctx, dao.TempQuota{
					StartTime: time.Now().Unix(),
					EndTime:   time.Now().Add(time.Hour * 24).Unix(),
					Key:       "temp-key-3",
					Amount:    10,
					UID:       3,
				})
				require.NoError(t, err)
				err = s.dao.AddQuota(ctx, "quota-key-3", dao.Quota{
					Amount: 5,
					UID:    3,
				})
				require.NoError(t, err)
			},
			after: func(ctx context.Context, t *testing.T) {
				list, err := s.dao.GetTempQuotaByUidAndTime(ctx, 3)
				require.NoError(t, err)
				require.Len(t, list, 1)
				assert.Equal(t, list[0].Amount, int64(0))

				quota, err := s.dao.GetQuotaByUid(ctx, 3)
				require.NoError(t, err)
				assert.Equal(t, int64(2), quota.Amount)
			},
		},
		{
			name:   "新用户,直接扣减为负数",
			uid:    4,
			amount: 2,
			key:    "deduct-key-4",
			before: func(context.Context, *testing.T) {
			},
			after: func(ctx context.Context, t *testing.T) {
				quota, err := s.dao.GetQuotaByUid(ctx, 4)
				require.NoError(t, err)
				assert.Equal(t, int64(-2), quota.Amount)
				assert.NotZero(t, quota.DebtStartTime)
			},
		},
		{
			name:   "用户充值了, 但是钱不够,触发 debtStartTime",
			uid:    5,
			amount: 13,
			key:    "deduct-key-5",
			before: func(ctx context.Context, t *testing.T) {
				err := s.dao.CreateTempQuota(ctx, dao.TempQuota{
					StartTime: time.Now().Unix(),
					EndTime:   time.Now().Add(time.Hour * 24).Unix(),
					Key:       "temp-key-5",
					Amount:    10,
					UID:       5,
				})
				require.NoError(t, err)
				err = s.dao.AddQuota(ctx, "quota-key-5", dao.Quota{
					UID:    5,
					Amount: 2,
				})
				require.NoError(t, err)
			},
			after: func(ctx context.Context, t *testing.T) {
				list, err := s.dao.GetTempQuotaByUidAndTime(ctx, 5)
				require.NoError(t, err)
				require.Len(t, list, 1)
				assert.Equal(t, list[0].Amount, int64(0))

				quota, err := s.dao.GetQuotaByUid(ctx, 5)
				require.NoError(t, err)
				assert.Equal(t, int64(-1), quota.Amount)
				assert.NotZero(t, quota.DebtStartTime)
			},
		},
		{
			name:   "用户amount 为0, 触发 debtStartTime",
			uid:    6,
			amount: 10,
			key:    "deduct-key-6",
			before: func(ctx context.Context, t *testing.T) {
				err := s.dao.CreateTempQuota(ctx, dao.TempQuota{
					StartTime: time.Now().Unix(),
					EndTime:   time.Now().Add(time.Hour * 24).Unix(),
					Key:       "temp-key-6",
					Amount:    0,
					UID:       6,
				})
				require.NoError(t, err)
				err = s.dao.AddQuota(ctx, "quota-key-6", dao.Quota{
					Amount: 0,
					UID:    6,
				})
				require.NoError(t, err)
			},
			after: func(ctx context.Context, t *testing.T) {
				list, err := s.dao.GetTempQuotaByUidAndTime(ctx, 6)
				require.NoError(t, err)
				require.Len(t, list, 1)
				assert.Equal(t, list[0].Amount, int64(0))

				quota, err := s.dao.GetQuotaByUid(ctx, 6)
				require.NoError(t, err)
				assert.Equal(t, int64(-10), quota.Amount)
				assert.NotZero(t, quota.DebtStartTime)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			tc.before(ctx, t)
			err := s.dao.Deduct(ctx, tc.uid, tc.amount, tc.key)
			require.NoError(t, err)
			tc.after(ctx, t)
		})
	}
}
