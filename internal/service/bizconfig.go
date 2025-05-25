// bizconfig.go 文件包含业务配置服务的接口定义和实现
// 提供了创建、获取、更新和删除业务配置的功能

package service

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

// BizConfigService 定义了处理业务配置的基本方法
type BizConfigService interface {
	// Create 创建一个新的业务配置
	// 参数 ctx 是上下文，用于控制请求的生命周期
	// 参数 config 是 domain.BizConfig 类型，表示要创建的配置
	// 返回值是创建后的 domain.BizConfig 对象和可能发生的错误
	Create(ctx context.Context, config domain.BizConfig) (domain.BizConfig, error)

	// GetByID 根据 ID 获取业务配置
	// 参数 ctx 是上下文，用于控制请求的生命周期
	// 参数 id 是 int64 类型，表示配置的 ID
	// 返回值是 domain.BizConfig 对象和可能发生的错误
	GetByID(ctx context.Context, id int64) (domain.BizConfig, error)

	// Update 更新业务配置
	// 参数 ctx 是上下文，用于控制请求的生命周期
	// 参数 config 是 domain.BizConfig 类型，表示要更新的配置
	// 返回值是可能发生的错误
	Update(ctx context.Context, config domain.BizConfig) error

	// Delete 根据 ID 删除业务配置
	// 参数 ctx 是上下文，用于控制请求的生命周期
	// 参数 id 是 string 类型，表示配置的 ID
	// 返回值是可能发生的错误
	Delete(ctx context.Context, id string) error
}

type bizConfigService struct {
	repo *repository.BizConfigRepository
}

// NewBizConfigService 创建一个新的 BizConfigService 实例
// 参数 repo 是 *repository.BizConfigRepository 类型，表示数据访问层
// 返回值是 BizConfigService 接口类型
func NewBizConfigService(repo *repository.BizConfigRepository) BizConfigService {
	return &bizConfigService{
		repo: repo,
	}
}

// Create 创建一个新的业务配置
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 req 是 domain.BizConfig 类型，表示客户端的请求
// 返回值是创建后的 domain.BizConfig 对象和可能发生的错误
func (s *bizConfigService) Create(ctx context.Context, req domain.BizConfig) (domain.BizConfig, error) {
	config := domain.BizConfig{
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Config:    req.Config,
	}

	created, err := s.repo.Create(ctx, config)
	if err != nil {
		return domain.BizConfig{}, err
	}

	return created, nil
}

// GetByID 根据 ID 获取业务配置
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 id 是 int64 类型，表示配置的 ID
// 返回值是 domain.BizConfig 对象和可能发生的错误
func (s *bizConfigService) GetByID(ctx context.Context, id int64) (domain.BizConfig, error) {
	return s.repo.GetByID(ctx, id)
}

// Update 更新业务配置
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 config 是 domain.BizConfig 类型，表示要更新的配置
// 返回值是可能发生的错误
func (s *bizConfigService) Update(ctx context.Context, config domain.BizConfig) error {
	return s.repo.Update(ctx, config)
}

// Delete 根据 ID 删除业务配置
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 id 是 string 类型，表示配置的 ID
// 返回值是可能发生的错误
func (s *bizConfigService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
