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

package dao

import (
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ekit/sqlx"
	"gorm.io/gorm/clause"

	"golang.org/x/net/context"
	"gorm.io/gorm"
)

type Biz struct {
	ID        int64                             `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string                            `gorm:"type:varchar(255)"`
	OwnerID   int64                             `gorm:"column:owner_id;type:bigint;not null"`
	OwnerType string                            `gorm:"column:owner_type;type:varchar(20);not null"`
	Config    sqlx.JsonColumn[domain.BizConfig] `gorm:"column:config;type:text"`
	Ctime     int64
	Utime     int64
}

func (Biz) TableName() string {
	return "biz_configs"
}

type BizDAO struct {
	db *gorm.DB
}

func NewBizConfigDAO(db *gorm.DB) *BizDAO {
	return &BizDAO{db: db}
}

func (d *BizDAO) Save(ctx context.Context, bc Biz) (int64, error) {
	now := time.Now().UnixMilli()
	bc.Ctime = now
	bc.Utime = now
	err := d.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"owner_id",
			"owner_type",
			"config",
			"utime"}),
	}).Create(&bc).Error
	return bc.ID, err
}

func (d *BizDAO) List(ctx context.Context, offset, limit int) ([]Biz, error) {
	var bc []Biz
	err := d.db.WithContext(ctx).Order("utime DESC").
		Offset(offset).Limit(limit).Find(&bc).Error
	return bc, err
}

func (d *BizDAO) Count(ctx context.Context) (int64, error) {
	var total int64
	err := d.db.WithContext(ctx).Model(&Biz{}).Count(&total).Error
	return total, err
}

func (d *BizDAO) GetByID(ctx context.Context, id int64) (Biz, error) {
	var bc Biz
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&bc).Error
	return bc, err
}
