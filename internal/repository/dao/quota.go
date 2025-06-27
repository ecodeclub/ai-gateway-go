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

package dao

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrNoAmount error = errors.New("余额不足")
)

type TempQuota struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;column:id"`
	UID       string `gorm:"column:uid"`
	Amount    int64  `gorm:"column:amount"`
	StartTime int64  `gorm:"column:start_time"`
	EndTime   int64  `gorm:"column:end_time"`
	Ctime     int64  `gorm:"column:ctime"`
	Utime     int64  `gorm:"column:utime"`
}

type Quota struct {
	ID            int64  `gorm:"primaryKey;autoIncrement;column:id"`
	UID           string `gorm:"column:uid"`
	Amount        int64  `gorm:"column:amount"`
	LastClearTime int64  `gorm:"column:last_clear_time"`
	Ctime         int64  `gorm:"column:ctime"`
	Utime         int64  `gorm:"column:utime"`
}

type QuotaDao struct {
	db *gorm.DB
}

func NewQuotaDao(db *gorm.DB) *QuotaDao {
	return &QuotaDao{db: db}
}

// CreateTempQuota 用来创建临时额度
func (dao *QuotaDao) CreateTempQuota(ctx context.Context, quota TempQuota) error {
	now := time.Now().Unix()
	quota.Ctime = now
	quota.Utime = now
	return dao.db.WithContext(ctx).Create(&quota).Error
}

// Create 用来创建对应的永久的额度
func (dao *QuotaDao) Create(ctx context.Context, quota Quota) error {
	now := time.Now().Unix()
	quota.Ctime = now
	quota.Ctime = now
	return dao.db.WithContext(ctx).Create(&quota).Error
}

func (dao *QuotaDao) GetQuotaByUid(ctx context.Context, uid string) error {
	return dao.db.WithContext(ctx).Where("uid = ?", uid).Error
}

func (dao *QuotaDao) GetTempQuotaByUid(ctx context.Context, uid string) error {
	return dao.db.WithContext(ctx).Where("uid = ?", uid).Error
}

// Deduct 扣减
func (dao *QuotaDao) Deduct(ctx context.Context, uid string, amount int64) error {
	now := time.Now().Unix()
	// 首先扣除temp 的, 然后扣除 quota的
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 可能存在多个时间段
		var tempQuotas []TempQuota
		err := tx.Where("uid = ? AND end_time >= ?", uid, now).
			Order("end_time ASC").
			Find(&tempQuotas).Error

		if err != nil {
			return err
		}

		for i := range tempQuotas {
			tq := tempQuotas[i]
			if amount <= 0 {
				break
			}
			deduct := int64(0)
			// 如果大于需要直接扣, 小于就直接扣完
			if tq.Amount >= amount {
				deduct = amount
				amount = 0
			} else {
				deduct = tq.Amount
				amount -= deduct
			}
			tq.Amount -= deduct
			tq.Utime = now
			// 然后更新
			err = tx.Model(tq).Select("amount", "utime").Updates(tq).Error
			if err != nil {
				return err
			}
		}

		var quota Quota
		err = tx.Where("uid = ?", uid).First(&quota).Error
		if err != nil {
			return err
		}

		// 临时额度扣减完毕
		if amount <= 0 {
			return nil
		}

		// 扣完了发现还不够扣的, 从 quota 中扣
		if quota.Amount < amount {
			return ErrNoAmount
		}
		quota.Amount -= amount
		quota.Utime = now
		quota.LastClearTime = now

		//更新
		err = tx.Model(&Quota{}).Updates(map[string]any{
			"amount":          quota.Amount,
			"utime":           quota.Utime,
			"last_clear_time": quota.LastClearTime,
		}).Error

		if err != nil {
			return ErrNoAmount
		}

		return nil
	})
	return err
}
