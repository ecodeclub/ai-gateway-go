package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Node struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Type     string `gorm:"column:type"`
	GraphID  int64  `gorm:"column:graph_id"`
	Status   string `gorm:"column:status"`
	Metadata string `gorm:"column:metadata"`
	Ctime    int64  `gorm:"column:ctime"`
	Utime    int64  `gorm:"column:utime"`
}

type Edge struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	GraphID  int64  `gorm:"column:graph_id"`
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

type GraphDao struct {
	db *gorm.DB
}

func NewGraphDao(db *gorm.DB) *GraphDao {
	return &GraphDao{db: db}
}

func (dao *GraphDao) Create(ctx context.Context, graph Graph) (int64, error) {
	now := time.Now().UnixMilli()
	graph.Ctime = now
	graph.Utime = now

	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(&graph).Error; err != nil {
			return err
		}
		for _, node := range graph.Nodes {
			node.Ctime = now
			node.Utime = now
			node.GraphID = graph.ID
			if err := tx.WithContext(ctx).Create(&node).Error; err != nil {
				return err
			}
		}
		return nil
	})

	return graph.ID, err
}

func (dao *GraphDao) UpdateGraphById(ctx context.Context, graph Graph) (int64, error) {
	now := time.Now().UnixMilli()
	graph.Utime = now

	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Where("id = ?", graph.ID).Updates(graph).Error
		if err != nil {
			return err
		}

		for _, node := range graph.Nodes {
			node.Utime = now
			err = tx.WithContext(ctx).Where("id = ? and graph_id = ?", node.ID, graph.ID).Updates(node).Error
			if err != nil {
				return err
			}
		}
		return nil
	})

	return graph.ID, err
}

func (dao *GraphDao) Save(ctx context.Context, graph Graph) (int64, error) {
	if graph.ID > 0 {
		return dao.UpdateGraphById(ctx, graph)
	} else {
		return dao.Create(ctx, graph)
	}
}

func (dao *GraphDao) Get(ctx context.Context, id int64) (Graph, error) {
	var graph Graph
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).First(&graph, id).Error; err != nil {
			return err
		}

		var nodes []Node
		if err := tx.WithContext(ctx).Where("graph_id = ?", id).Find(&nodes).Error; err != nil {
			return err
		}
		graph.Nodes = nodes

		return nil
	})

	return graph, err
}
