package dao

type Node struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	GraphID  int64  `gorm:"column:GraphID"`
	Type     string `gorm:"column:type"`
	Status   string `gorm:"column:status"`
	Metadata string `gorm:"column:metadata"`
	Ctime    int64  `gorm:"column:ctime"`
	Utime    int64  `gorm:"column:utime"`
}

type Edge struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	GraphID  int64  `gorm:"column:GraphID"`
	SourceID int64  `gorm:"column:source_id;index:idx_source_target"`
	TargetID int64  `gorm:"column:target_id;index:idx_source_target"`
	Metadata string `gorm:"column:metadata;"`
	Ctime    int64  `gorm:"column:ctime"`
	Utime    int64  `gorm:"column:utime"`
}

type Graph struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Edges    []Edge `gorm:"column:foreignKey:GraphID"`
	Nodes    []Node `gorm:"column:foreignKey:GraphID"`
	MetaData string `gorm:"column:metadata"`
	Ctime    int64  `gorm:"column:ctime"`
	Utime    int64  `gorm:"column:utime"`
}
