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
	"context"
	"database/sql"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"gorm.io/gorm/clause"

	"gorm.io/gorm"
)

type InvocationConfig struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string `gorm:"column:name"`
	BizID       int64  `gorm:"column:biz_id"`
	Description string `gorm:"column:description"`
	Ctime       int64  `gorm:"column:ctime"`
	Utime       int64  `gorm:"column:utime"`
}

type InvocationConfigVersion struct {
	ID           int64            `gorm:"column:id;primaryKey;autoIncrement"`
	InvID        int64            `gorm:"column:inv_id;index"`
	ModelID      int64            `gorm:"column:model_id"`
	Version      string           `gorm:"column:version;type:varchar(255)"`
	Prompt       string           `gorm:"column:prompt"`
	SystemPrompt string           `gorm:"column:system_prompt"`
	JSONSchema   sql.Null[string] `gorm:"column:json_schema;type:longText;comment:'结构化数据的JSONSchema'"`
	Temperature  float32          `gorm:"column:temperature"`
	TopP         float32          `gorm:"column:top_p"`
	MaxTokens    int              `gorm:"column:max_tokens"`
	Status       string           `gorm:"column:status;"`
	Ctime        int64            `gorm:"column:ctime"`
	Utime        int64            `gorm:"column:utime"`
}

type InvocationConfigDAO struct {
	db *gorm.DB
}

func NewInvocationConfigDAO(db *gorm.DB) *InvocationConfigDAO {
	return &InvocationConfigDAO{db: db}
}

func (p *InvocationConfigDAO) Save(ctx context.Context, cfg InvocationConfig) (int64, error) {
	now := time.Now().UnixMilli()
	cfg.Ctime, cfg.Utime = now, now
	err := p.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"name", "biz_id", "description", "utime"}),
	}).Create(&cfg).Error
	return cfg.ID, err
}

func (p *InvocationConfigDAO) List(ctx context.Context, offset, limit int) ([]InvocationConfig, error) {
	var res []InvocationConfig
	err := p.db.WithContext(ctx).Offset(offset).Limit(limit).Order("utime DESC").Find(&res).Error
	return res, err
}

func (p *InvocationConfigDAO) Count(ctx context.Context) (int, error) {
	var res int64
	err := p.db.WithContext(ctx).Model(&InvocationConfig{}).Count(&res).Error
	return int(res), err
}

func (p *InvocationConfigDAO) GetByID(ctx context.Context, id int64) (InvocationConfig, error) {
	var res InvocationConfig
	err := p.db.WithContext(ctx).Where("id = ?", id).First(&res).Error
	return res, err
}

func (p *InvocationConfigDAO) SaveVersion(ctx context.Context, version InvocationConfigVersion) (int64, error) {
	now := time.Now().UnixMilli()
	version.Utime = now
	version.Ctime = now
	err := p.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{
			"inv_id",
			"model_id",
			"version",
			"prompt",
			"system_prompt",
			"json_schema",
			"temperature",
			"top_p",
			"max_tokens",
			"status",
			"utime"}),
	}).Save(&version).Error
	return version.ID, err
}

func (p *InvocationConfigDAO) ListVersions(ctx context.Context, invID int64, offset, limit int) ([]InvocationConfigVersion, error) {
	var res []InvocationConfigVersion
	err := p.db.WithContext(ctx).Where("inv_id = ?", invID).
		Order("utime DESC").Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (p *InvocationConfigDAO) CountVersions(ctx context.Context, invID int64) (int, error) {
	var res int64
	err := p.db.WithContext(ctx).Model(&InvocationConfigVersion{}).
		Where("inv_id = ?", invID).Count(&res).Error
	return int(res), err
}

func (p *InvocationConfigDAO) GetVersionByD(ctx context.Context, id int64) (InvocationConfigVersion, error) {
	var res InvocationConfigVersion
	err := p.db.WithContext(ctx).Model(&InvocationConfigVersion{}).Where("id = ?", id).First(&res).Error
	return res, err
}

func (p *InvocationConfigDAO) ActivateVersion(ctx context.Context, id int64) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		var version InvocationConfigVersion
		err := tx.Where("id = ?", id).First(&version).Error
		if err != nil {
			return err
		}
		// 把之前激活的取消
		err = tx.Model(&InvocationConfigVersion{}).Where("inv_id = ? AND status = ?", version.InvID, domain.InvocationCfgVersionStatusActive.String()).
			Updates(map[string]any{
				"status": domain.InvocationCfgVersionStatusDraft.String(),
				"utime":  now,
			}).Error
		if err != nil {
			return err
		}
		// 激活当前版本
		return tx.Model(&InvocationConfigVersion{}).Where("id = ?", id).Updates(map[string]any{
			"status": domain.InvocationCfgVersionStatusActive.String(),
			"utime":  now,
		}).Error
	})
}
