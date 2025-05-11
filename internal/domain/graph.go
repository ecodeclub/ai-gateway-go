package domain

import (
	"github.com/ecodeclub/ekit"
)

type Step struct {
	ID       int64
	GraphID  int64
	Type     string
	Status   string
	Metadata ekit.AnyValue // 用于存储扩展属性
}

type Edge struct {
	ID       int64
	GraphID  int64
	SourceID int64
	TargetID int64
	Weight   float64
	Metadata ekit.AnyValue // 用于存储边的扩展属性
}

type Plan struct {
	ID       int64
	Steps    []Step
	Edges    []Edge
	Metadata ekit.AnyValue
}
