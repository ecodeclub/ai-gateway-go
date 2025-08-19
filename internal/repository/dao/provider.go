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
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Provider struct {
	ID     int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name   string `gorm:"column:name"`
	APIKey string `gorm:"colum:api_key"`
	Ctime  int64  `gorm:"colum:ctime"`
	Utime  int64  `gorm:"colum:utime"`
}

func (Provider) TableName() string {
	return "providers"
}

type Model struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string `gorm:"column:name"`
	Pid         int64  `gorm:"column:pid"`
	InputPrice  int64  `gorm:"column:input_price"`
	OutputPrice int64  `gorm:"column:output_price"`
	PriceMode   string `gorm:"colum:price_mode"`
	Ctime       int64  `gorm:"colum:ctime"`
	Utime       int64  `gorm:"column:utime"`
}

func (Model) TableName() string {
	return "models"
}

type ProviderDAO struct {
	db *gorm.DB
}

func NewProviderDAO(db *gorm.DB) *ProviderDAO {
	return &ProviderDAO{db: db}
}

func (d *ProviderDAO) SaveProvider(ctx context.Context, provider Provider) (int64, error) {
	now := time.Now().UnixMilli()
	provider.Utime = now
	provider.Ctime = now
	err := d.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{
			"api_key",
			"name",
			"utime",
		}),
	}).Create(&provider).Error
	return provider.ID, err
}

func (d *ProviderDAO) ListProviders(ctx context.Context, offset, limit int) ([]Provider, error) {
	var providers []Provider
	err := d.db.WithContext(ctx).Model(&Provider{}).Order("utime DESC").
		Offset(offset).Limit(limit).Find(&providers).Error
	return providers, err
}

func (d *ProviderDAO) CountProviders(ctx context.Context) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&Provider{}).Count(&count).Error
	return count, err
}

func (d *ProviderDAO) GetProvider(ctx context.Context, id int64) (Provider, error) {
	var provider Provider
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&provider).Error
	return provider, err
}

func (d *ProviderDAO) GetModelsByPid(ctx context.Context, pid int64) ([]Model, error) {
	var model []Model
	err := d.db.WithContext(ctx).Model(&Model{}).Where("pid = ?", pid).Order("id DESC").Find(&model).Error
	return model, err
}

func (d *ProviderDAO) SaveModel(ctx context.Context, model Model) (int64, error) {
	now := time.Now().Unix()
	model.Ctime = now
	model.Utime = now
	err := d.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{
			"pid",
			"name",
			"input_price",
			"output_price",
			"price_mode",
			"utime",
		}),
	}).Create(&model).Error
	return model.ID, err
}

func (d *ProviderDAO) GetModel(ctx context.Context, id int64) (Model, error) {
	var model Model
	err := d.db.WithContext(ctx).Model(&model).Where("id = ?", id).First(&model).Error
	return model, err
}
