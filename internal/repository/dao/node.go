package dao

type Node struct {
	ID       int64          `gorm:"column:id;primaryKey;autoIncrement"`
	Type     string         `gorm:"column:type"`
	Statue   string         `gorm:"column:statue"`
	Metadata map[string]any `gorm:"column:metadata;type:json"`
	Ctime    int64          `gorm:"column:ctime"`
	Utime    int64          `gorm:"column:utime"`
}

type Edge struct {
	ID       int64          `gorm:"column:id;primaryKey;autoIncrement"`
	SourceID int64          `gorm:"column:source_id;index:idx_source_target"`
	TargetID int64          `gorm:"column:target_id;index:idx_source_target"`
	Metadata map[string]any `gorm:"column:metadata;type:json"`
	Ctime    int64          `gorm:"column:ctime"`
	Utime    int64          `gorm:"column:utime"`
}
