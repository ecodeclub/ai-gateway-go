package node

// BaseNode 基础节点实现
type BaseNode struct {
	params     map[string]any
	successors map[string]Node
}

func (b *BaseNode) SetParams(params map[string]any) {
	b.params = params
}

func (b *BaseNode) AddSuccessor(action string, node Node) {
	if b.successors == nil {
		b.successors = make(map[string]Node)
	}
	b.successors[action] = node
}

func (b *BaseNode) GetSuccessor() map[string]Node {
	return b.successors
}

func (b *BaseNode) Clone() Node {
	return &BaseNode{
		params:     CopyMap(b.params),
		successors: CopyMap(b.successors),
	}
}

func CopyMap[K comparable, V any](m map[K]V) map[K]V {
	cpy := make(map[K]V, len(m))
	for k, v := range m {
		cpy[k] = v
	}

	return cpy
}
