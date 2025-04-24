package node

import (
	"context"
	"fmt"
)

type SequentialNode struct {
	*BaseNode
	plan       Plan
	maxRetries int
	waitMs     int
}

func NewSequentialNode(plan Plan) *SequentialNode {
	return &SequentialNode{
		BaseNode:   &BaseNode{},
		plan:       plan,
		maxRetries: 1,
	}
}

func (s *SequentialNode) RunSequential(ctx context.Context, shared SharedContext) (string, error) {
	prepResult, err := s.plan.Prepare(ctx, shared)
	if err != nil {
		return "", fmt.Errorf("prepare failed %w", err)
	}

	var (
		result  string
		execErr error
	)

	// 当失败之后进行重试
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		result, execErr = s.plan.Execute(ctx, prepResult)
		if execErr == nil {
			break
		}
	}
	if execErr != nil {
		return "", fmt.Errorf("execute failed after %d attempts: %w", s.maxRetries, execErr)
	}
	if err := s.plan.Post(ctx, shared, prepResult, result); err != nil {
		return "", fmt.Errorf("post processing failed: %w", err)
	}

	return result, nil
}
