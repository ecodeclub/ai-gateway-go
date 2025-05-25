package dao

import "gorm.io/gorm"

// InitTables 初始化数据库表结构
// 参数:
//
//	db: 指向gorm.DB的数据库连接指针
//
// 返回值:
//
//	error: 初始化过程中发生的错误
//
// 注意：当前仅初始化BizConfig表，可根据需要扩展其他表
func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&BizConfig{})
}
