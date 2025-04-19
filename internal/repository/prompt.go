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

func (p *PromptRepo) Delete(ctx context.Context, id, versionID int64) error {
	return p.dao.Delete(ctx, id, versionID)
}

func (p *PromptRepo) Update(ctx context.Context, value domain.Prompt) error {
	if value.ID > 0 {
		err := p.dao.UpdatePrompt(ctx, dao.Prompt{
			ID:          value.ID,
			Name:        value.Name,
			Description: value.Description,
		})
		if err != nil {
			return err
		}
	}
	if len(value.Versions) > 0 {
		err := p.dao.UpdateVersion(ctx, dao.PromptVersion{
			ID:            value.Versions[0].ID,
			Label:         value.Versions[0].Label,
			Content:       value.Versions[0].Content,
			SystemContent: value.Versions[0].SystemContent,
			Temperature:   value.Versions[0].Temperature,
			TopN:          value.Versions[0].TopN,
			MaxTokens:     value.Versions[0].MaxTokens,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

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

func (p *PromptRepo) Publish(ctx context.Context, id int64, versionID int64, label string) error {
	return p.dao.Publish(ctx, id, versionID, label)
}

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
