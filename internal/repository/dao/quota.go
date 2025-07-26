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

	"github.com/ecodeclub/ai-gateway-go/errs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TempQuota struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;column:id"`
	UID       int64  `gorm:"column:uid"`
	Key       string `gorm:"column:key;uniqueIndex;type:varchar(256)"`
	Amount    int64  `gorm:"column:amount"`
	StartTime int64  `gorm:"column:start_time"`
	EndTime   int64  `gorm:"column:end_time"`
	Ctime     int64  `gorm:"column:ctime"`
	Utime     int64  `gorm:"column:utime"`
}

func (TempQuota) TableName() string {
	return "temp_quotas"
}

type QuotaRecord struct {
	ID     int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Uid    int64  `gorm:"column:uid;index"`
	Key    string `gorm:"column:key;uniqueIndex;type:varchar(256)"`
	Amount int64  `gorm:"column:amount"`
	Ctime  int64  `gorm:"column:ctime"`
	Utime  int64  `gorm:"column:utime"`
}

func (QuotaRecord) TableName() string {
	return "quota_records"
}

type Quota struct {
	ID            int64  `gorm:"primaryKey;autoIncrement;column:id"`
	UID           int64  `gorm:"column:uid"`
	Key           string `gorm:"column:key;uniqueIndex;type:varchar(256)"`
	Amount        int64  `gorm:"column:amount"`
	LastClearTime int64  `gorm:"column:last_clear_time"`
	Ctime         int64  `gorm:"column:ctime"`
	Utime         int64  `gorm:"column:utime"`
}

func (Quota) TableName() string {
	return "quotas"
}

type QuotaDao struct {
	db *gorm.DB
}

func NewQuotaDao(db *gorm.DB) *QuotaDao {
	return &QuotaDao{db: db}
}

func (dao *QuotaDao) CreateTempQuota(ctx context.Context, quota TempQuota) error {
	now := time.Now().Unix()
	quota.Ctime = now
	quota.Utime = now
	return dao.db.WithContext(ctx).Create(&quota).Error
}

func (dao *QuotaDao) AddQuota(ctx context.Context, quota Quota) error {
	now := time.Now().Unix()
	quota.Utime = now

	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().Unix()
		record := QuotaRecord{
			Key:    quota.Key,
			Uid:    quota.UID,
			Amount: quota.Amount,
			Ctime:  now,
			Utime:  now,
		}
		err := tx.Create(&record).Error
		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"amount": gorm.Expr("amount + ?", quota.Amount),
				"utime":  now,
			}),
		}).Create(&quota).Error
	})
}

func (dao *QuotaDao) GetQuotaByUid(ctx context.Context, uid int64) (Quota, error) {
	var quota Quota
	err := dao.db.WithContext(ctx).
		Where("uid = ? and end_time >= ?", uid).
		First(&quota).Error
	if err != nil {
		return Quota{}, err
	}
	return quota, nil
}

func (dao *QuotaDao) GetTempQuotaByUidAndTime(ctx context.Context, uid int64) ([]TempQuota, error) {
	now := time.Now().Unix()
	var quota []TempQuota
	err := dao.db.WithContext(ctx).
		Where("uid = ? and end_time >= ?", uid, now).
		Order("end_time ASC").
		Find(&quota).Error
	if err != nil {
		return nil, err
	}
	return quota, nil
}

func (dao *QuotaDao) Deduct(ctx context.Context, uid int64, amount int64, key string) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().Unix()
		record := QuotaRecord{
			Key:    key,
			Uid:    uid,
			Amount: amount,
			Ctime:  now,
			Utime:  now,
		}
		err := tx.Create(&record).Error
		if err != nil {
			return err
		}
		return dao.deduct(tx, uid, amount, now)
	})
}

func (dao *QuotaDao) deduct(tx *gorm.DB, uid int64, amount int64, now int64) error {
	deductAmount := amount
	for {
		var quota TempQuota
		err := tx.Where("amount > ? and uid = ?", 0, uid).First(&quota).Error
		if err != nil {
			// 表示找不到可以扣减的temp
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return err
		}
		deductAmount = min(deductAmount, quota.Amount)
		result := tx.Model(&TempQuota{}).Where("amount > ? and uid = ?", 0, uid).Updates(map[string]any{
			"amount": gorm.Expr("amount - ?", deductAmount),
			"utime":  now,
		})
		if result.Error != nil {
			return result.Error
		}
		// 并发问题, 直接下一个
		if result.RowsAffected == 0 {
			continue
		}
		// 表示扣减完毕
		amount -= deductAmount
		if amount <= 0 {
			return nil
		}
	}
	// 从主额度扣
	result := tx.Model(&Quota{}).
		Where("uid = ? AND amount >= ?", uid, deductAmount).
		Updates(map[string]any{
			"amount": gorm.Expr("amount - ?", deductAmount),
			"utime":  now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errs.ErrInsufficientBalance
	}
	return nil
}

func InitQuotaTable(db *gorm.DB) error {
	return db.AutoMigrate(&Quota{}, &TempQuota{}, &QuotaRecord{})
}
