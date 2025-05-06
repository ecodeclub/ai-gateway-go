package domain

import (
	"github.com/ecodeclub/ekit"
)

type Node struct {
	ID       int64
	GraphID  int64
	Type     string
	Status   string
	Metadata ekit.AnyValue // 用于存储扩展属性
}

type Edge struct {
	Id       int64
	GraphID  int64
	SourceID int64
	TargetID int64
	Weight   float64
	Metadata ekit.AnyValue // 用于存储边的扩展属性
}

type Graph struct {
	Id       int64
	Nodes    []Node
	Edges    []Edge
	Metadata ekit.AnyValue
}
