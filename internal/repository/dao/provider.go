package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProviderDao struct {
	db *gorm.DB
}

func NewProvider(db *gorm.DB) *ProviderDao {
	return &ProviderDao{db: db}
}

func (d *ProviderDao) SaveProvider(ctx context.Context, provider Provider) (int64, error) {
	err := d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"api_key": provider.APIKey,
			"name":    provider.Name,
			"utime":   provider.Ctime,
		}),
	}).Create(&provider).Error
	return provider.Id, err
}

func (d *ProviderDao) SaveModel(ctx context.Context, model Model) (int64, error) {
	now := time.Now().Unix()
	model.Ctime = now
	model.Utime = now

	err := d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"pid":          model.Pid,
			"input_price":  model.InputPrice,
			"output_price": model.OutputPrice,
			"price_mode":   model.PriceMode,
			"utime":        model.Ctime,
		}),
	}).Create(&model).Error
	return model.Id, err
}

func (d *ProviderDao) GetModels(ctx context.Context, pid int64) ([]Model, error) {
	var model []Model
	err := d.db.WithContext(ctx).Model(&Model{}).
		Where("pid = ?", pid).
		Find(&model).Error

	if err != nil {
		return nil, err
	}
	return model, nil
}

func (d *ProviderDao) GetAllProviders(ctx context.Context) ([]Provider, error) {
	var providers []Provider
	err := d.db.WithContext(ctx).Model(&Provider{}).Find(&providers).Error
	return providers, err
}

type Provider struct {
	Id     int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name   string `gorm:"column:name"`
	APIKey string `gorm:"colum:api_key"`
	Ctime  int64  `gorm:"colum:ctime"`
	Utime  int64  `gorm:"colum:utime"`
}

func (Provider) TableName() string {
	return "providers"
}

type Model struct {
	Id          int64  `gorm:"column:id;primaryKey;autoIncrement"`
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
