package dao

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type PromptDAO struct {
	db *gorm.DB
}

func NewPromptDAO(db *gorm.DB) *PromptDAO {
	return &PromptDAO{db: db}
}

func (p *PromptDAO) Create(ctx context.Context, prompt Prompt, version PromptVersion) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		prompt.Ctime, prompt.Utime = now, now
		version.Ctime, version.Utime = now, now
		err := tx.Create(&prompt).Error
		if err != nil {
			return err
		}
		version.PromptID = prompt.ID
		return tx.Create(&version).Error
	})
}

func (p *PromptDAO) FindByID(ctx context.Context, id int64) (Prompt, []PromptVersion, error) {
	var res Prompt
	err := p.db.WithContext(ctx).Where("id = ?", id).First(&res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return res, nil, nil
	}
	var versions []PromptVersion
	p.db.WithContext(ctx).Model(&PromptVersion{}).Where("prompt_id = ?", res.ID).First(&versions)
	return res, versions, err
}

func (p *PromptDAO) UpdatePrompt(ctx context.Context, value Prompt) error {
	// 更新非零值
	value.Utime = time.Now().UnixMilli()
	return p.db.WithContext(ctx).Model(&Prompt{}).Where("id = ?", value.ID).Updates(value).Error
}

func (p *PromptDAO) UpdateVersion(ctx context.Context, value PromptVersion) error {
	// 更新非零值
	value.Utime = time.Now().UnixMilli()
	return p.db.WithContext(ctx).Model(&PromptVersion{}).Where("id = ?", value.ID).Updates(value).Error
}

func (p *PromptDAO) Delete(ctx context.Context, id int64) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		err := tx.Model(&Prompt{}).Where("id = ?", id).Updates(map[string]any{
			"status": 0,
			"utime":  now,
		}).Error
		if err != nil {
			return err
		}
		return tx.Model(&PromptVersion{}).Where("prompt_id = ?", id).Updates(map[string]any{
			"status": 0,
			"utime":  now,
		}).Error
	})
}

func (p *PromptDAO) DeleteVersion(ctx context.Context, versionID int64) error {
	return p.db.WithContext(ctx).Model(&PromptVersion{}).Where("id = ?", versionID).Updates(map[string]any{
		"status": 0,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (p *PromptDAO) UpdateActiveVersion(ctx context.Context, versionID int64, label string) error {
	var id int64
	err := p.db.WithContext(ctx).Model(&PromptVersion{}).Where("id = ?", versionID).Select("prompt_id").First(&id).Error
	if err != nil {
		return err
	}

	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		err := tx.Model(&Prompt{}).Where("id = ?", id).Updates(map[string]any{
			"active_version": versionID,
			"utime":          now,
		}).Error
		if err != nil {
			return err
		}
		return tx.Model(&PromptVersion{}).Where("id = ?", versionID).Updates(map[string]any{
			"label": label,
			"utime": now,
		}).Error
	})
}

func (p *PromptDAO) InsertVersion(ctx context.Context, version PromptVersion) error {
	now := time.Now().UnixMilli()
	version.Ctime = now
	version.Utime = now
	return p.db.WithContext(ctx).Create(&version).Error
}

func (p *PromptDAO) GetByVersionID(ctx context.Context, versionID int64) (PromptVersion, error) {
	var res PromptVersion
	err := p.db.WithContext(ctx).Model(&PromptVersion{}).Where("id = ?", versionID).First(&res).Error
	return res, err
}

type Prompt struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name          string `gorm:"column:name"`
	Owner         int64  `gorm:"column:owner;index:idx_owner_owner_type"`
	OwnerType     string `gorm:"column:owner_type;type:ENUM('personal','organization');index:idx_owner_owner_type"`
	ActiveVersion int64  `gorm:"column:active_version"`
	Description   string `gorm:"column:description"`
	Status        uint8  `gorm:"column:status;default:1"`
	Ctime         int64  `gorm:"column:ctime"`
	Utime         int64  `gorm:"column:utime"`
}

func (Prompt) TableName() string {
	return "prompts"
}

type PromptVersion struct {
	ID            int64   `gorm:"column:id;primaryKey;autoIncrement"`
	PromptID      int64   `gorm:"column:prompt_id;index"`
	Label         string  `gorm:"column:label"`
	Content       string  `gorm:"column:content"`
	SystemContent string  `gorm:"column:system_content"`
	Temperature   float32 `gorm:"column:temperature"`
	TopN          float32 `gorm:"column:top_n"`
	MaxTokens     int     `gorm:"column:max_tokens"`
	Status        uint8   `gorm:"column:status;default:1"`
	Ctime         int64   `gorm:"column:ctime"`
	Utime         int64   `gorm:"column:utime"`
}

func (PromptVersion) TableName() string {
	return "prompt_versions"
}

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&Prompt{}, &PromptVersion{})
}
