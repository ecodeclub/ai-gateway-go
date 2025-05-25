package repository

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"time"
)

// PromptRepo 是提示模板的仓库实现，负责在领域模型和数据访问层之间进行转换。
type PromptRepo struct {
	dao *dao.PromptDAO // 数据访问对象，用于操作数据库
}

// NewPromptRepo 创建一个新的PromptRepo实例。
// 参数:
//
//	dao: 数据访问对象实例
//
// 返回值:
//
//	*PromptRepo: 初始化后的PromptRepo实例
func NewPromptRepo(dao *dao.PromptDAO) *PromptRepo {
	return &PromptRepo{dao: dao}
}

// Create 创建一个新的提示及其初始版本。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	prompt: 要创建的提示基础信息
//	version: 与提示关联的初始版本信息
//
// 返回值:
//
//	error: 执行过程中发生的错误，如果成功则为nil
func (p *PromptRepo) Create(ctx context.Context, prompt domain.Prompt, version domain.PromptVersion) error {
	dPrompt := dao.Prompt{
		Name:        prompt.Name,
		Owner:       prompt.Owner,
		OwnerType:   string(prompt.OwnerType),
		Description: prompt.Description,
	}
	dVersion := dao.PromptVersion{
		Content:       version.Content,
		SystemContent: version.SystemContent,
		Temperature:   version.Temperature,
		TopN:          version.TopN,
		MaxTokens:     version.MaxTokens,
	}
	return p.dao.Create(ctx, dPrompt, dVersion)
}

// Get 根据ID获取完整的提示信息（包括所有版本）。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 要查询的提示ID
//
// 返回值:
//
//	domain.Prompt: 查询到的提示信息
//	error: 执行过程中发生的错误
func (p *PromptRepo) Get(ctx context.Context, id int64) (domain.Prompt, error) {
	prompt, dVersions, err := p.dao.FindByID(ctx, id)
	if err != nil {
		return domain.Prompt{}, err
	}
	versions := make([]domain.PromptVersion, 0, len(dVersions))
	for _, v := range dVersions {
		versions = append(versions, domain.PromptVersion{
			ID:            v.ID,
			Label:         v.Label,
			Content:       v.Content,
			SystemContent: v.SystemContent,
			Temperature:   v.Temperature,
			TopN:          v.TopN,
			MaxTokens:     v.MaxTokens,
			Status:        v.Status,
			Ctime:         time.UnixMilli(v.Ctime),
			Utime:         time.UnixMilli(v.Utime),
		})
	}
	return domain.Prompt{
		ID:            prompt.ID,
		Name:          prompt.Name,
		Owner:         prompt.Owner,
		OwnerType:     domain.OwnerType(prompt.OwnerType),
		ActiveVersion: prompt.ActiveVersion,
		Versions:      versions,
		Description:   prompt.Description,
		Ctime:         time.UnixMilli(prompt.Ctime),
		Utime:         time.UnixMilli(prompt.Utime),
	}, nil
}

// Delete 在事务中软删除提示及其所有版本。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 要删除的提示ID
//
// 返回值:
//
//	error: 执行过程中发生的错误
func (p *PromptRepo) Delete(ctx context.Context, id int64) error {
	return p.dao.Delete(ctx, id)
}

// DeleteVersion 软删除指定版本。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	versionID: 要删除的版本ID
//
// 返回值:
//
//	error: 执行过程中发生的错误
func (p *PromptRepo) DeleteVersion(ctx context.Context, versionID int64) error {
	return p.dao.DeleteVersion(ctx, versionID)
}

// UpdateInfo 更新提示的基本信息（名称和描述）。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	value: 包含更新数据的提示对象
//
// 返回值:
//
//	error: 执行过程中发生的错误
func (p *PromptRepo) UpdateInfo(ctx context.Context, value domain.Prompt) error {
	return p.dao.UpdatePrompt(ctx, dao.Prompt{
		ID:          value.ID,
		Name:        value.Name,
		Description: value.Description,
	})
}

// UpdateVersion 更新指定版本的信息。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	value: 包含更新数据的版本对象
//
// 返回值:
//
//	error: 执行过程中发生的错误
func (p *PromptRepo) UpdateVersion(ctx context.Context, value domain.PromptVersion) error {
	return p.dao.UpdateVersion(ctx, dao.PromptVersion{
		ID:            value.ID,
		Content:       value.Content,
		SystemContent: value.SystemContent,
		Temperature:   value.Temperature,
		TopN:          value.TopN,
		MaxTokens:     value.MaxTokens,
	})
}

// UpdateActiveVersion 更新提示的激活版本。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	versionID: 要激活的新版本ID
//	label: 新版本标签
//
// 返回值:
//
//	error: 执行过程中发生的错误
func (p *PromptRepo) UpdateActiveVersion(ctx context.Context, versionID int64, label string) error {
	return p.dao.UpdateActiveVersion(ctx, versionID, label)
}

// InsertVersion 插入新的版本记录。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 提示ID
//	version: 要插入的版本对象
//
// 返回值:
//
//	error: 执行过程中发生的错误
func (p *PromptRepo) InsertVersion(ctx context.Context, id int64, version domain.PromptVersion) error {
	return p.dao.InsertVersion(ctx, dao.PromptVersion{
		PromptID:      id,
		Content:       version.Content,
		SystemContent: version.SystemContent,
		Temperature:   version.Temperature,
		TopN:          version.TopN,
		MaxTokens:     version.MaxTokens,
	})
}

// GetByVersionID 根据版本ID查询版本详情。
// 参数:
//
//	ctx: 上下文对象用于控制请求生命周期
//	id: 要查询的版本ID
//
// 返回值:
//
//	domain.Prompt: 查询到的提示信息（仅包含请求的版本）
//	error: 执行过程中发生的错误
func (p *PromptRepo) GetByVersionID(ctx context.Context, id int64) (domain.Prompt, error) {
	res, err := p.dao.GetByVersionID(ctx, id)
	if err != nil {
		return domain.Prompt{}, err
	}
	version := domain.PromptVersion{
		ID:            res.ID,
		Label:         res.Label,
		Content:       res.Content,
		SystemContent: res.SystemContent,
		Temperature:   res.Temperature,
		TopN:          res.TopN,
		MaxTokens:     res.MaxTokens,
		Status:        res.Status,
		Ctime:         time.UnixMilli(res.Ctime),
		Utime:         time.UnixMilli(res.Utime),
	}
	return domain.Prompt{
		ID:       res.PromptID,
		Versions: []domain.PromptVersion{version},
	}, nil
}
