// Copyright 2025 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
}

func (Edge) TableName() string {
	return "edges"
}

type Graph struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Metadata string `gorm:"column:metadata"`
	Ctime    int64  `gorm:"column:ctime"`
	Utime    int64  `gorm:"column:utime"`
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
