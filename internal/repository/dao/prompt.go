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

// NewPromptDAO 创建一个新的 PromptDAO 实例。
// 参数:
//
//	db: 指向 gorm.DB 的数据库连接指针
//
// 返回值:
//
//	*PromptDAO: 初始化后的 PromptDAO 结构体指针
func NewPromptDAO(db *gorm.DB) *PromptDAO {
	return &PromptDAO{db: db}
}

/*
 * Create 在事务中创建提示及其初始版本。
 * 参数:
 *   ctx: 上下文对象用于控制请求生命周期
 *   prompt: 要创建的提示基础信息
 *   version: 与提示关联的初始版本信息
 * 返回值:
 *   error: 执行过程中发生的错误，如果成功则为nil
 */
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

/*
 * FindByID 根据ID查询完整提示信息（包含所有版本）。
 * 参数:
 *   ctx: 上下文对象用于控制请求生命周期
 *   id: 要查询的提示ID
 * 返回值:
 *   Prompt: 查询到的提示信息
 *   []PromptVersion: 关联的所有版本信息
 *   error: 执行过程中发生的错误
 */
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

/*
 * UpdatePrompt 更新提示的基本信息。
 * 参数:
 *   ctx: 上下文对象用于控制请求生命周期
 *   value: 包含更新数据的提示对象
 * 返回值:
 *   error: 执行过程中发生的错误
 */
func (p *PromptDAO) UpdatePrompt(ctx context.Context, value Prompt) error {
	// 更新非零值
	value.Utime = time.Now().UnixMilli()
	return p.db.WithContext(ctx).Model(&Prompt{}).Where("id = ?", value.ID).Updates(value).Error
}

/*
 * UpdateVersion 更新指定版本的信息。
 * 参数:
 *   ctx: 上下文对象用于控制请求生命周期
 *   value: 包含更新数据的版本对象
 * 返回值:
 *   error: 执行过程中发生的错误
 */
func (p *PromptDAO) UpdateVersion(ctx context.Context, value PromptVersion) error {
	// 更新非零值
	value.Utime = time.Now().UnixMilli()
	return p.db.WithContext(ctx).Model(&PromptVersion{}).Where("id = ?", value.ID).Updates(value).Error
}

/*
 * Delete 在事务中软删除提示及其所有版本。
 * 参数:
 *   ctx: 上下文对象用于控制请求生命周期
 *   id: 要删除的提示ID
 * 返回值:
 *   error: 执行过程中发生的错误
 */
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

/*
 * DeleteVersion 软删除指定版本。
 * 参数:
 *   ctx: 上下文对象用于控制请求生命周期
 *   versionID: 要删除的版本ID
 * 返回值:
 *   error: 执行过程中发生的错误
 */
func (p *PromptDAO) DeleteVersion(ctx context.Context, versionID int64) error {
	return p.db.WithContext(ctx).Model(&PromptVersion{}).Where("id = ?", versionID).Updates(map[string]any{
		"status": 0,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

/*
 * UpdateActiveVersion 更新提示的激活版本。
 * 参数:
 *   ctx: 上下文对象用于控制请求生命周期
 *   versionID: 要激活的新版本ID
 *   label: 新版本标签
 * 返回值:
 *   error: 执行过程中发生的错误
 */
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

/*
 * InsertVersion 插入新的版本记录。
 * 参数:
 *   ctx: 上下文对象用于控制请求生命周期
 *   version: 要插入的版本对象
 * 返回值:
 *   error: 执行过程中发生的错误
 */
func (p *PromptDAO) InsertVersion(ctx context.Context, version PromptVersion) error {
	now := time.Now().UnixMilli()
	version.Ctime = now
	version.Utime = now
	return p.db.WithContext(ctx).Create(&version).Error
}

/*
 * GetByVersionID 根据版本ID查询版本详情。
 * 参数:
 *   ctx: 上下文对象用于控制请求生命周期
 *   versionID: 要查询的版本ID
 * 返回值:
 *   PromptVersion: 查询到的版本信息
 *   error: 执行过程中发生的错误
 */
func (p *PromptDAO) GetByVersionID(ctx context.Context, versionID int64) (PromptVersion, error) {
	var res PromptVersion
	err := p.db.WithContext(ctx).Model(&PromptVersion{}).Where("id = ?", versionID).First(&res).Error
	return res, err
}

// Prompt 结构体定义
// 表示一个提示模板的基本信息
type Prompt struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement"`                                                // 提示的唯一标识符
	Name          string `gorm:"column:name"`                                                                       // 提示名称
	Owner         int64  `gorm:"column:owner;index:idx_owner_owner_type"`                                           // 所有者ID
	OwnerType     string `gorm:"column:owner_type;type:ENUM('personal','organization');index:idx_owner_owner_type"` // 所有者类型：个人或组织
	ActiveVersion int64  `gorm:"column:active_version"`                                                             // 当前激活的版本ID
	Description   string `gorm:"column:description"`                                                                // 提示描述
	Status        uint8  `gorm:"column:status;default:1"`                                                           // 状态：0表示删除，1表示有效
	Ctime         int64  `gorm:"column:ctime"`                                                                      // 创建时间戳（毫秒）
	Utime         int64  `gorm:"column:utime"`                                                                      // 最后更新时间戳（毫秒）
}

func (Prompt) TableName() string {
	return "prompts"
}

// PromptVersion 结构体定义
// 表示提示模板的具体版本信息
type PromptVersion struct {
	ID            int64   `gorm:"column:id;primaryKey;autoIncrement"` // 版本ID
	PromptID      int64   `gorm:"column:prompt_id;index"`             // 关联的提示ID
	Label         string  `gorm:"column:label"`                       // 版本标签
	Content       string  `gorm:"column:content"`                     // 主要内容
	SystemContent string  `gorm:"column:system_content"`              // 系统消息内容
	Temperature   float32 `gorm:"column:temperature"`                 // 模型温度参数
	TopN          float32 `gorm:"column:top_n"`                       // Top N采样参数
	MaxTokens     int     `gorm:"column:max_tokens"`                  // 最大生成token数
	Status        uint8   `gorm:"column:status;default:1"`            // 状态：0表示删除，1表示有效
	Ctime         int64   `gorm:"column:ctime"`                       // 创建时间戳（毫秒）
	Utime         int64   `gorm:"column:utime"`                       // 最后更新时间戳（毫秒）
}

func (PromptVersion) TableName() string {
	return "prompt_versions"
}

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&Prompt{}, &PromptVersion{})
}
