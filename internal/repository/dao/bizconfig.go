package dao

import (
	"errors"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"time"
)

var ErrBizConfigNotFound = errors.New("biz config not found")

type BizConfig struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement"`
	OwnerID   int64  `gorm:"column:owner_id;type:bigint;not null"`
	OwnerType string `gorm:"column:owner_type;type:varchar(20);not null"`
	Token     string `gorm:"column:token;type:varchar(64);not null"`
	Config    string `gorm:"column:config;type:text"`
	CreatedAt int64
	UpdatedAt int64
}

func (BizConfig) TableName() string {
	return "biz_configs"
}

type BizConfigDAO struct {
	db *gorm.DB
}

func NewBizConfigDAO(db *gorm.DB) *BizConfigDAO {
	return &BizConfigDAO{db: db}
}

// Create 创建记录，返回创建后的对象（含 ID 和 Token）
func (d *BizConfigDAO) Insert(ctx context.Context, bc *BizConfig) (*BizConfig, error) {
	now := time.Now().UnixMilli()
	bc.CreatedAt = now
	bc.UpdatedAt = now
	err := d.db.WithContext(ctx).Create(bc).Error
	if err != nil {
		return nil, err
	}
	return bc, nil
}

// GetByID 根据ID查询
func (d *BizConfigDAO) GetByID(ctx context.Context, id int64) (*BizConfig, error) {
	var bc BizConfig
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&bc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrBizConfigNotFound
	}
	return &bc, err
}

// Update 更新记录
func (d *BizConfigDAO) Update(ctx context.Context, bc *BizConfig) error {
	return d.db.WithContext(ctx).Save(bc).Error
}

// Delete 删除记录
func (d *BizConfigDAO) Delete(ctx context.Context, id string) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&BizConfig{}).Error
}
