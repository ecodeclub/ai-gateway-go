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

func (s *PromptService) Add(ctx context.Context, value domain.Prompt) error {
	return s.repo.Add(ctx, value)
}

func (s *PromptService) Get(ctx context.Context, id int64) (domain.Prompt, error) {
	return s.repo.Get(ctx, id)
}

func (s *PromptService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *PromptService) Update(ctx context.Context, id int64, name string, content string, description string) error {
	return s.repo.Update(ctx, domain.Prompt{
		ID:          id,
		Name:        name,
		Content:     content,
		Description: description,
	})
}
