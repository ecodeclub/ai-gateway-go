package service

import (
	"context"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

// NodeService 是一个结构体，包含一个 NodeRepo 的指针
type NodeService struct {
	repo *repository.NodeRepo
}

// NewGraphService 创建一个新的 NodeService 实例
// 参数 repo 是 *repository.NodeRepo 类型，表示数据访问层
// 返回值是 *NodeService 类型
func NewGraphService(repo *repository.NodeRepo) *NodeService {
	return &NodeService{repo: repo}
}

// GetGraph 根据 ID 获取图
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 id 是 int64 类型，表示图的 ID
// 返回值是 domain.Graph 对象和可能发生的错误
func (svc *NodeService) GetGraph(ctx context.Context, id int64) (domain.Graph, error) {
	return svc.repo.GetGraph(ctx, id)
}

// SaveGraph 保存图
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 graph 是 domain.Graph 类型，表示要保存的图
// 返回值是保存后的 ID 和可能发生的错误
func (svc *NodeService) SaveGraph(ctx context.Context, graph domain.Graph) (int64, error) {
	return svc.repo.SaveGraph(ctx, graph)
}

// SaveNode 保存节点
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 step 是 domain.Node 类型，表示要保存的节点
// 返回值是保存后的 ID 和可能发生的错误
func (svc *NodeService) SaveNode(ctx context.Context, step domain.Node) (int64, error) {
	return svc.repo.SaveNode(ctx, step)
}

// SaveEdge 保存边
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 edge 是 domain.Edge 类型，表示要保存的边
// 返回值是保存后的 ID 和可能发生的错误
func (svc *NodeService) SaveEdge(ctx context.Context, edge domain.Edge) (int64, error) {
	return svc.repo.SaveEdge(ctx, edge)
}

// DeleteEdge 删除边
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 id 是 int64 类型，表示要删除的边的 ID
// 返回值是可能发生的错误
func (svc *NodeService) DeleteEdge(ctx context.Context, id int64) error {
	return svc.repo.DeleteEdge(ctx, id)
}

// DeleteNode 删除节点
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 id 是 int64 类型，表示要删除的节点的 ID
// 返回值是可能发生的错误
func (svc *NodeService) DeleteNode(ctx context.Context, id int64) error {
	return svc.repo.DeleteNode(ctx, id)
}

// DeleteGraph 删除图
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 id 是 int64 类型，表示要删除的图的 ID
// 返回值是可能发生的错误
func (svc *NodeService) DeleteGraph(ctx context.Context, id int64) error {
	return svc.repo.DeleteGraph(ctx, id)
}
