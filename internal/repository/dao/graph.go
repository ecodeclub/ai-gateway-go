package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Node 结构体定义
// 表示图中的一个节点
type Node struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"` // 节点的唯一标识符
	Type     string `gorm:"column:type"`                        // 节点类型
	GraphID  int64  `gorm:"column:graph_id;index"`              // 所属图的ID
	Status   string `gorm:"column:status"`                      // 节点状态
	Metadata string `gorm:"column:metadata"`                    // 节点元数据，存储为JSON格式字符串
	Ctime    int64  `gorm:"column:ctime"`                       // 创建时间戳（毫秒）
	Utime    int64  `gorm:"column:utime"`                       // 最后更新时间戳（毫秒）
}

func (Node) TableName() string {
	return "nodes"
}

// Edge 结构体定义
// 表示图中的一个边（连接）
type Edge struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`       // 边的唯一标识符
	GraphID  int64  `gorm:"column:graph_id;index"`                    // 所属图的ID
	SourceID int64  `gorm:"column:source_id;index:idx_source_target"` // 源节点ID
	TargetID int64  `gorm:"column:target_id;index:idx_source_target"` // 目标节点ID
	Metadata string `gorm:"column:metadata;"`                         // 边的元数据，存储为JSON格式字符串
	Ctime    int64  `gorm:"column:ctime"`                             // 创建时间戳（毫秒）
	Utime    int64  `gorm:"column:utime"`                             // 最后更新时间戳（毫秒）
}

func (Edge) TableName() string {
	return "edges"
}

// Graph 结构体定义
// 表示一个完整的图结构
type Graph struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"` // 图的唯一标识符
	Metadata string `gorm:"column:metadata"`                    // 图的元数据，存储为JSON格式字符串
	Ctime    int64  `gorm:"column:ctime"`                       // 创建时间戳（毫秒）
	Utime    int64  `gorm:"column:utime"`                       // 最后更新时间戳（毫秒）
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
	return dao.db.WithContext(ctx).Delete(&Node{}, id).Error
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
		DoUpdates: clause.AssignmentColumns([]string{"graph_id", "target_id", "source_id", "metadata", "utime"}),
	}).Create(&edge).Error
	return edge.ID, err
}

func (dao *GraphDAO) DeleteEdge(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Delete(&Edge{}, id).Error
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
		DoUpdates: clause.AssignmentColumns([]string{"utime", "metadata"}),
	}).Create(&graph).Error
	return graph.ID, err
}

func (dao *GraphDAO) DeleteGraph(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&Graph{}).Where("id = ?", id).Delete(&Graph{}).Error
		if err != nil {
			return err
		}

		err = tx.Model(&Node{}).Where("graph_id = ?", id).Delete(&Node{}).Error
		if err != nil {
			return err
		}

		err = tx.Model(&Edge{}).Where("graph_id = ?", id).Delete(&Edge{}).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (dao *GraphDAO) GetNodes(ctx context.Context, id int64) ([]Node, error) {
	var nodes []Node
	err := dao.db.WithContext(ctx).Where("graph_id = ?", id).Find(&nodes).Error
	if err != nil {
		return nodes, err
	}
	return nodes, nil
}

func (dao *GraphDAO) GetEdges(ctx context.Context, id int64) ([]Edge, error) {
	var edges []Edge
	err := dao.db.WithContext(ctx).Where("graph_id = ?", id).Find(&edges).Error
	if err != nil {
		return edges, err
	}
	return edges, nil
}

func (dao *GraphDAO) GetGraph(ctx context.Context, id int64) (Graph, error) {
	var graph Graph
	if err := dao.db.WithContext(ctx).First(&graph, id).Error; err != nil {
		return Graph{}, err
	}
	return graph, nil
}

func InitGraphTable(db *gorm.DB) error {
	return db.AutoMigrate(&Graph{}, &Edge{}, &Node{})
}
