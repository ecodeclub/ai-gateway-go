package repository

import (
	"context"
	"errors"
	"github.com/ecodeclub/ai-gateway-go/errs"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"time"
)

// BizConfigRepository 是业务配置的仓库实现，负责在领域模型和数据访问层之间进行转换。
type BizConfigRepository struct {
	dao *dao.BizConfigDAO // 数据访问对象，用于操作数据库
}

// NewBizConfigRepository 创建一个新的BizConfigRepository实例。
// 参数:
//
//	dao: 数据访问对象实例
//
// 返回值:
//
//	*BizConfigRepository: 初始化后的BizConfigRepository实例
func NewBizConfigRepository(dao *dao.BizConfigDAO) *BizConfigRepository {
	return &BizConfigRepository{dao: dao}
}

// Create 将一个新的业务配置记录插入数据库。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	config: 要插入的业务配置对象
//
// 返回值:
//
//	domain.BizConfig: 插入成功的业务配置对象
//	error: 插入过程中发生的错误
func (r *BizConfigRepository) Create(ctx context.Context, config domain.BizConfig) (domain.BizConfig, error) {
	daoBC, err := r.dao.Insert(ctx, toDAOConfig(config))
	if err != nil {
		return domain.BizConfig{}, err
	}
	return fromDAOConfig(daoBC), nil
}

// GetByID 根据ID查询业务配置。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 要查询的配置ID
//
// 返回值:
//
//	domain.BizConfig: 查询到的业务配置对象
//	error: 查询过程中发生的错误
func (r *BizConfigRepository) GetByID(ctx context.Context, id int64) (domain.BizConfig, error) {
	bc, err := r.dao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrBizConfigNotFound) {
			return domain.BizConfig{}, errs.ErrBizConfigNotFound
		}
		return domain.BizConfig{}, err
	}
	return fromDAOConfig(bc), nil
}

// Update 更新现有的业务配置记录。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	config: 包含更新数据的业务配置对象
//
// 返回值:
//
//	error: 更新过程中发生的错误
func (r *BizConfigRepository) Update(ctx context.Context, config domain.BizConfig) error {
	return r.dao.Update(ctx, toDAOConfig(config))
}

// Delete 根据ID删除业务配置记录。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 要删除的配置ID
//
// 返回值:
//
//	error: 删除过程中发生的错误
func (r *BizConfigRepository) Delete(ctx context.Context, id string) error {
	return r.dao.Delete(ctx, id)
}

// toDAOConfig 将领域模型的BizConfig转换为DAO层的BizConfig。
// 参数:
//
//	config: 领域模型的BizConfig对象
//
// 返回值:
//
//	*dao.BizConfig: DAO层的BizConfig对象
func toDAOConfig(config domain.BizConfig) *dao.BizConfig {
	return &dao.BizConfig{
		ID:        config.ID,
		OwnerID:   config.OwnerID,
		OwnerType: config.OwnerType,
		Config:    config.Config,
		Ctime:     config.Ctime.UnixMilli(),
		Utime:     config.Utime.UnixMilli(),
	}
}

// fromDAOConfig 将DAO层的BizConfig转换为领域模型的BizConfig。
// 参数:
//
//	bc: DAO层的BizConfig对象
//
// 返回值:
//
//	domain.BizConfig: 领域模型的BizConfig对象
func fromDAOConfig(bc dao.BizConfig) domain.BizConfig {
	return domain.BizConfig{
		ID:        bc.ID,
		OwnerID:   bc.OwnerID,
		OwnerType: bc.OwnerType,
		Config:    bc.Config,
		Ctime:     time.UnixMilli(bc.Ctime),
		Utime:     time.UnixMilli(bc.Utime),
	}
}
