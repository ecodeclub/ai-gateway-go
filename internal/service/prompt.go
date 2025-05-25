package service

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

// PromptService 是一个结构体，包含一个 PromptRepo 的指针
type PromptService struct {
	repo *repository.PromptRepo
}

// NewPromptService 创建一个新的 PromptService 实例
// 参数 repo 是 *repository.PromptRepo 类型，表示数据访问层
// 返回值是 *PromptService 类型
func NewPromptService(repo *repository.PromptRepo) *PromptService {
	return &PromptService{repo: repo}
}

// Add 添加提示信息
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 prompt 是 domain.Prompt 类型，表示要添加的提示
// 参数 version 是 domain.PromptVersion 类型，表示提示的版本
// 返回值是可能发生的错误
func (s *PromptService) Add(ctx context.Context, prompt domain.Prompt, version domain.PromptVersion) error {
	return s.repo.Create(ctx, prompt, version)
}

// Get 根据 ID 获取提示信息
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 id 是 int64 类型，表示提示的 ID
// 返回值是 domain.Prompt 对象和可能发生的错误
func (s *PromptService) Get(ctx context.Context, id int64) (domain.Prompt, error) {
	return s.repo.Get(ctx, id)
}

// Delete 根据 ID 删除提示信息
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 id 是 int64 类型，表示要删除的提示的 ID
// 返回值是可能发生的错误
func (s *PromptService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// DeleteVersion 删除提示版本
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 versionID 是 int64 类型，表示要删除的提示版本的 ID
// 返回值是可能发生的错误
func (s *PromptService) DeleteVersion(ctx context.Context, versionID int64) error {
	return s.repo.DeleteVersion(ctx, versionID)
}

// UpdateInfo 更新提示信息
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 prompt 是 domain.Prompt 类型，表示要更新的提示信息
// 返回值是可能发生的错误
func (s *PromptService) UpdateInfo(ctx context.Context, prompt domain.Prompt) error {
	return s.repo.UpdateInfo(ctx, prompt)
}

// UpdateVersion 更新提示版本
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 version 是 domain.PromptVersion 类型，表示要更新的提示版本信息
// 返回值是可能发生的错误
func (s *PromptService) UpdateVersion(ctx context.Context, version domain.PromptVersion) error {
	return s.repo.UpdateVersion(ctx, version)
}

// Publish 发布提示版本
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 versionID 是 int64 类型，表示要发布的提示版本的 ID
// 参数 label 是 string 类型，表示发布的标签
// 返回值是可能发生的错误
func (s *PromptService) Publish(ctx context.Context, versionID int64, label string) error {
	return s.repo.UpdateActiveVersion(ctx, versionID, label)
}

// Fork 分叉提示版本
// 参数 ctx 是上下文，用于控制请求的生命周期
// 参数 versionID 是 int64 类型，表示要分叉的提示版本的 ID
// 返回值是可能发生的错误
func (s *PromptService) Fork(ctx context.Context, versionID int64) error {
	prompt, err := s.repo.GetByVersionID(ctx, versionID)
	if err != nil {
		return err
	}
	newVersion := domain.PromptVersion{
		Content:       prompt.Versions[0].Content,
		SystemContent: prompt.Versions[0].SystemContent,
		Temperature:   prompt.Versions[0].Temperature,
		TopN:          prompt.Versions[0].TopN,
		MaxTokens:     prompt.Versions[0].MaxTokens,
	}
	return s.repo.InsertVersion(ctx, prompt.ID, newVersion)
}
