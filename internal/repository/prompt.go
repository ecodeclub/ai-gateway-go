package repository

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"time"
)

type PromptRepo struct {
	dao *dao.PromptDAO
}

func NewPromptRepo(dao *dao.PromptDAO) *PromptRepo {
	return &PromptRepo{dao: dao}
}

func (p *PromptRepo) Add(ctx context.Context, prompt domain.Prompt) error {
	return p.dao.Create(ctx, dao.Prompt{
		Name:        prompt.Name,
		Owner:       prompt.Owner,
		OwnerType:   string(prompt.OwnerType),
		Content:     prompt.Content,
		Description: prompt.Description,
	})
}

func (p *PromptRepo) Get(ctx context.Context, id int64) (domain.Prompt, error) {
	prompt, err := p.dao.FindByID(ctx, id)
	if err != nil {
		return domain.Prompt{}, err
	}
	return domain.Prompt{
		ID:          prompt.ID,
		Name:        prompt.Name,
		Owner:       prompt.Owner,
		OwnerType:   domain.OwnerType(prompt.OwnerType),
		Content:     prompt.Content,
		Description: prompt.Description,
		Ctime:       time.UnixMilli(prompt.Ctime),
		Utime:       time.UnixMilli(prompt.Utime),
	}, nil
}

func (p *PromptRepo) Delete(ctx context.Context, id int64) error {
	return p.dao.Delete(ctx, id)
}

func (p *PromptRepo) Update(ctx context.Context, value domain.Prompt) error {
	return p.dao.Update(ctx, dao.Prompt{
		ID:          value.ID,
		Name:        value.Name,
		Content:     value.Content,
		Description: value.Description,
	})
}
