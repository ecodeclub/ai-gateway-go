package repository

import (
	"context"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
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

func (p *PromptRepo) Delete(ctx context.Context, id int64) error {
	return p.dao.Delete(ctx, id)
}

func (p *PromptRepo) DeleteVersion(ctx context.Context, versionID int64) error {
	return p.dao.DeleteVersion(ctx, versionID)
}

func (p *PromptRepo) UpdateInfo(ctx context.Context, value domain.Prompt) error {
	return p.dao.UpdatePrompt(ctx, dao.Prompt{
		ID:          value.ID,
		Name:        value.Name,
		Description: value.Description,
	})
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

func (p *PromptRepo) UpdateActiveVersion(ctx context.Context, versionID int64, label string) error {
	return p.dao.UpdateActiveVersion(ctx, versionID, label)
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
