package dao

type Node struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Type     string `gorm:"column:type"`
	Status   string `gorm:"column:status"`
	Metadata string `gorm:"column:metadata"`
	Ctime    int64  `gorm:"column:ctime"`
	Utime    int64  `gorm:"column:utime"`
}

type Edge struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	SourceID int64  `gorm:"column:source_id;index:idx_source_target"`
	TargetID int64  `gorm:"column:target_id;index:idx_source_target"`
	Metadata string `gorm:"column:metadata;"`
	Ctime    int64  `gorm:"column:ctime"`
	Utime    int64  `gorm:"column:utime"`
}

type Graph struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Edges    []Edge `gorm:"-"`
	Nodes    []Node `gorm:"-"`
	Metadata string `gorm:"column:metadata"`
	Ctime    int64  `gorm:"column:ctime"`
	Utime    int64  `gorm:"column:utime"`
}
