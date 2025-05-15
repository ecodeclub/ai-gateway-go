package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Node struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Type     string `gorm:"column:type"`
	GraphID  int64  `gorm:"column:graph_id;index"`
	Status   string `gorm:"column:status"`
	Metadata string `gorm:"column:metadata"`
	Ctime    int64  `gorm:"column:ctime"`
	Utime    int64  `gorm:"column:utime"`
	Deleted  uint8  `gorm:"column:deleted;default:1"`
}

func (Node) TableName() string {
	return "nodes"
}

type Edge struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	GraphID  int64  `gorm:"column:graph_id;index"`
	SourceID int64  `gorm:"column:source_id;index:idx_source_target"`
	TargetID int64  `gorm:"column:target_id;index:idx_source_target"`
	Metadata string `gorm:"column:metadata;"`
	Ctime    int64  `gorm:"column:ctime"`
	Utime    int64  `gorm:"column:utime"`
	Deleted  uint8  `gorm:"column:deleted;default:1"`
}

func (Edge) TableName() string {
	return "edges"
}

type Graph struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Edges    []Edge `gorm:"-"`
	Nodes    []Node `gorm:"-"`
	Metadata string `gorm:"column:metadata"`
	Ctime    int64  `gorm:"column:ctime"`
	Utime    int64  `gorm:"column:utime"`
	Deleted  uint8  `gorm:"column:deleted;default:1"`
}

func (Graph) TableName() string {
	return "graphs"
}

type GraphDAO struct {
	db *gorm.DB
}

func NewGraphDAO(db *gorm.DB) *GraphDAO {
	return &GraphDAO{db: db}
}

func (dao *GraphDAO) SaveNode(ctx context.Context, node Node) (int64, error) {
	now := time.Now().UnixMilli()
	if node.ID > 0 {
		node.Utime = now
	} else {
		node.Ctime = now
		node.Utime = now
	}

	err := dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"graph_id", "type", "status", "metadata", "utime"}),
	}).Create(&node).Error
	return node.ID, err
}

func (dao *GraphDAO) DeleteNode(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	err := dao.db.WithContext(ctx).Model(&Node{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted": 0,
		"utime":   now,
	}).Error
	return err
}

func (dao *GraphDAO) SaveEdge(ctx context.Context, edge Edge) (int64, error) {
	now := time.Now().UnixMilli()
	if edge.ID > 0 {
		edge.Utime = now
	} else {
		edge.Ctime = now
		edge.Utime = now
	}

	err := dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"graph_id", "target_id", "source_id", "utime"}),
	}).Create(&edge).Error
	return edge.ID, err
}

func (dao *GraphDAO) DeleteEdge(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()

	err := dao.db.WithContext(ctx).Model(&Edge{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted": 0,
		"utime":   now,
	}).Error

	return err
}

func (dao *GraphDAO) SaveGraph(ctx context.Context, graph Graph) (int64, error) {
	now := time.Now().UnixMilli()
	if graph.ID > 0 {
		graph.Utime = now
	} else {
		graph.Ctime = now
		graph.Utime = now
	}

	err := dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&graph).Error
	return graph.ID, err
}

func (dao *GraphDAO) DeleteGraph(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&Graph{}).Where("id = ?", id).Updates(map[string]interface{}{
			"deleted": 0,
			"utime":   now,
		}).Error
		if err != nil {
			return err
		}

		err = tx.Model(&Edge{}).Where("graph_id = ?", id).Updates(map[string]interface{}{
			"deleted": 0,
			"utime":   now,
		}).Error

		if err != nil {
			return err
		}

		err = tx.Model(&Node{}).Where("graph_id = ?", id).Updates(map[string]interface{}{
			"deleted": 0,
			"utime":   now,
		}).Error

		if err != nil {
			return err
		}
		return nil
	})
}

func (dao *GraphDAO) Get(ctx context.Context, id int64) (Graph, error) {
	var graph Graph
	if err := dao.db.WithContext(ctx).First(&graph, id).Error; err != nil {
		return Graph{}, err
	}

	if err := dao.db.WithContext(ctx).Where("id = ? and deleted != ?", id, 0).First(&graph).Error; err != nil {
		return Graph{}, err
	}

	var nodes []Node
	if err := dao.db.WithContext(ctx).Where("graph_id = ? and deleted != ?", id, 0).Find(&nodes).Error; err != nil {
		return Graph{}, err
	}
	graph.Nodes = nodes

	var edges []Edge
	if err := dao.db.WithContext(ctx).Where("graph_id = ? and deleted != ?", id, 0).Find(&edges).Error; err != nil {
		return Graph{}, err
	}
	graph.Edges = edges
	return graph, nil
}

func InitGraphTable(db *gorm.DB) error {
	return db.AutoMigrate(&Graph{}, &Edge{}, &Node{})
}
