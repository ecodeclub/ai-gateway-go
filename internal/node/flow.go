package node

import "context"

type FlowController struct {
	startNode Node
}

func NewFlowController(node Node) *FlowController {
	return &FlowController{startNode: node}
}

func (f *FlowController) Execute(ctx context.Context, shared SharedContext) error {
	current := f.startNode
	var lastAction string

	for current != nil {
		// 如果当前是串行节点
		if sn, ok := current.(SerialNode); ok {
			action, err := sn.RunSequential(ctx, shared)
			if err != nil {
				return err
			}
			lastAction = action
		}

		// 得到它的下一个节点
		successors := current.GetSuccessor()
		if nextNode, exists := successors[lastAction]; exists {
			current = nextNode.Clone()
		} else {
			current = nil
		}
	}

	return nil
}
