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

func (p *PromptDAO) Create(ctx context.Context, name, biz, pattern, desc string) error {
	now := time.Now().UnixMilli()
	value := Prompt{
		Name:        name,
		Biz:         biz,
		Pattern:     pattern,
		Description: desc,
		Ctime:       now,
		Utime:       now,
	}
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

func (p *PromptDAO) Update(ctx context.Context, id int64, name, pattern, desc string) error {
	now := time.Now().UnixMilli()
	m := map[string]any{
		"utime": now,
	}
	if name != "" {
		m["name"] = name
	}
	if pattern != "" {
		m["pattern"] = pattern
	}
	if desc != "" {
		m["description"] = desc
	}
	return p.db.WithContext(ctx).Model(&Prompt{}).Where("id = ?", id).Updates(m).Error
}

func (p *PromptDAO) Delete(ctx context.Context, id int64) error {
	return p.db.WithContext(ctx).Model(&Prompt{}).Where("id = ?", id).Updates(map[string]any{
		"status": 0,
		"utime":  time.Now().UnixMilli(),
	}).Error
}
