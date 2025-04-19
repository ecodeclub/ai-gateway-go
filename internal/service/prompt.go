package service

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

type PromptService struct {
	repo *repository.PromptRepo
}

func NewPromptService(repo *repository.PromptRepo) *PromptService {
	return &PromptService{repo: repo}
}

func (s *PromptService) Add(ctx context.Context, prompt domain.Prompt, version domain.PromptVersion) error {
	return s.repo.Create(ctx, prompt, version)
}

func (s *PromptService) Get(ctx context.Context, id int64) (domain.Prompt, error) {
	return s.repo.Get(ctx, id)
}

func (s *PromptService) Delete(ctx context.Context, id int64, versionID int64) error {
	return s.repo.Delete(ctx, id, versionID)
}

func (s *PromptService) Update(ctx context.Context, prompt domain.Prompt) error {
	return s.repo.Update(ctx, prompt)
}

func (s *PromptService) Publish(ctx context.Context, id int64, versionID int64, label string) error {
	return s.repo.Publish(ctx, id, versionID, label)
}

func (s *PromptService) Fork(ctx context.Context, id int64, version domain.PromptVersion) error {
	return s.repo.InsertVersion(ctx, id, version)
}
