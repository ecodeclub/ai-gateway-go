package node

import "context"

type (
	SharedContext map[string]any

	// Node 节点基础接口
	Node interface {
		SetParams(params map[string]any)
		AddSuccessor(action string, node Node)
		GetSuccessor() map[string]Node
		Clone() Node
	}

	// Plan 表示是一个执行计划, 这个计划是用户去规划的
	Plan interface {
		Prepare(ctx context.Context, shared SharedContext) (any, error)
		Execute(ctx context.Context, prepResult any) (string, error)
		Post(ctx context.Context, shared SharedContext, prepResult any, execResult string) error
	}

	// SerialNode 串行节点
	SerialNode interface {
		Node
		Plan
		RunSequential(ctx context.Context, shared SharedContext) (string, error)
	}

	// ParallelNode 并行节点
	ParallelNode interface {
		Node
		Plan
		RunParallelNode(ctx context.Context, shared SharedContext) (<-chan string, <-chan string)
	}
)
