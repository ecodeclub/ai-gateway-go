package dao

type Prompt struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string `gorm:"column:name"`
	Biz         string `gorm:"column:biz;uniqueIndex"`
	Pattern     string `gorm:"column:pattern"`
	Description string `gorm:"column:description"`
	Status      uint8  `gorm:"column:status;default:1"`
	Ctime       int64  `gorm:"column:ctime"`
	Utime       int64  `gorm:"column:utime"`
}

func (Prompt) TableName() string {
	return "prompts"
}
