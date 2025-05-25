package dao

import (
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"time"
)

// BizConfig 结构体定义
// 表示业务配置信息，用于存储不同所有者的配置数据
type BizConfig struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement"`          // 配置的唯一标识符
	OwnerID   int64  `gorm:"column:owner_id;type:bigint;not null"`        // 所有者ID（如用户ID或组织ID）
	OwnerType string `gorm:"column:owner_type;type:varchar(20);not null"` // 所有者类型：如"user"或"organization"
	Config    string `gorm:"column:config;type:text"`                     // 配置内容，通常存储JSON格式数据
	Ctime     int64  `gorm:"column:ctime"`                                // 创建时间戳（毫秒）
	Utime     int64  `gorm:"column:utime"`                                // 最后更新时间戳（毫秒）
}

func (BizConfig) TableName() string {
	return "biz_configs"
}

// BizConfigDAO 是用于操作BizConfig数据库的DAO（数据访问对象）结构体。
type BizConfigDAO struct {
	db *gorm.DB // 指向GORM数据库连接的指针
}

// NewBizConfigDAO 创建一个新的BizConfigDAO实例。
func NewBizConfigDAO(db *gorm.DB) *BizConfigDAO {
	return &BizConfigDAO{db: db}
}

// Insert 将一个新的业务配置记录插入数据库，并设置其创建和更新时间。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	bc: 要插入的BizConfig对象指针
//
// 返回值:
//
//	BizConfig: 插入成功的BizConfig对象
//	error: 插入过程中发生的错误
func (d *BizConfigDAO) Insert(ctx context.Context, bc *BizConfig) (BizConfig, error) {
	now := time.Now().UnixMilli()
	bc.Ctime = now
	bc.Utime = now
	err := d.db.WithContext(ctx).Create(bc).Error
	if err != nil {
		return BizConfig{}, err
	}
	return *bc, nil
}

// GetByID 根据ID查询业务配置。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 要查询的配置ID
//
// 返回值:
//
//	BizConfig: 查询到的BizConfig对象
//	error: 查询过程中发生的错误
func (d *BizConfigDAO) GetByID(ctx context.Context, id int64) (BizConfig, error) {
	var bc BizConfig
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&bc).Error
	return bc, err
}

// Update 更新现有的业务配置记录。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	bc: 包含更新数据的BizConfig对象指针
//
// 返回值:
//
//	error: 更新过程中发生的错误
func (d *BizConfigDAO) Update(ctx context.Context, bc *BizConfig) error {
	bc.Utime = time.Now().UnixMilli()
	return d.db.WithContext(ctx).Save(bc).Error
}

// Delete 根据ID删除业务配置记录。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 要删除的配置ID
//
// 返回值:
//
//	error: 删除过程中发生的错误
func (d *BizConfigDAO) Delete(ctx context.Context, id string) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&BizConfig{}).Error
}
