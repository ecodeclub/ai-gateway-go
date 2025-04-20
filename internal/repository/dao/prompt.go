package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

type PromptDAO struct {
	db *gorm.DB
}

func NewPromptDAO(db *gorm.DB) *PromptDAO {
	return &PromptDAO{db: db}
}

func (p *PromptDAO) Create(ctx context.Context, value Prompt) error {
	return p.db.WithContext(ctx).Create(&value).Error
}

func (p *PromptDAO) FindByID(ctx context.Context, id int64) (Prompt, error) {
	var res Prompt
	err := p.db.WithContext(ctx).Where("id = ?", id).First(&res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return res, nil
	}
	return res, err
}

func (p *PromptDAO) Update(ctx context.Context, value Prompt) error {
	// 更新非零值
	return p.db.WithContext(ctx).Model(&Prompt{}).Where("id = ?", value.ID).Updates(value).Error
}

func (p *PromptDAO) Delete(ctx context.Context, id int64) error {
	return p.db.WithContext(ctx).Model(&Prompt{}).Where("id = ?", id).Updates(map[string]any{
		"status": 0,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

type Prompt struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string `gorm:"column:name"`
	Owner       int64  `gorm:"column:owner;index:idx_owner_owner_type"`
	OwnerType   string `gorm:"column:owner_type;type:ENUM('personal','organization');index:idx_owner_owner_type"`
	Content     string `gorm:"column:content"`
	Description string `gorm:"column:description"`
	Status      uint8  `gorm:"column:status;default:1"`
	Ctime       int64  `gorm:"column:ctime"`
	Utime       int64  `gorm:"column:utime"`
}

func (Prompt) TableName() string {
	return "prompts"
}
