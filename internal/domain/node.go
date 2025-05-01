package domain

type Node struct {
	ID       int64
	Type     string
	Status   string
	Metadata map[string]any // 用于存储扩展属性
}

type Edge struct {
	SourceID int64
	TargetID int64
	Weight   float64
	Metadata map[string]any // 用于存储边的扩展属性
}

type Graph struct {
	Nodes []Node
	Edges []Edge
}
