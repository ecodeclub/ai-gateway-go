package web

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
)

type PromptVO struct {
	ID            int64             `json:"id"`
	Name          string            `json:"name"`
	Owner         int64             `json:"owner"`
	OwnerType     string            `json:"owner_type"`
	ActiveVersion int64             `json:"active_version"`
	Versions      []PromptVersionVO `json:"versions"`
	Description   string            `json:"description"`
	CreateTime    int64             `json:"create_time"`
	UpdateTime    int64             `json:"update_time"`
}

type PromptVersionVO struct {
	ID            int64   `json:"id"`
	Label         string  `json:"label"`
	Content       string  `json:"content"`
	SystemContent string  `json:"system_content"`
	Temperature   float32 `json:"temperature"`
	TopN          float32 `json:"top_n"`
	MaxTokens     int     `json:"max_tokens"`
	Status        uint8   `json:"status"`
	CreateTime    int64   `json:"ctime"`
	UpdateTime    int64   `json:"utime"`
}

func newPromptVO(p domain.Prompt) PromptVO {
	versions := make([]PromptVersionVO, len(p.Versions))
	for i, v := range p.Versions {
		versions[i] = PromptVersionVO{
			ID:            v.ID,
			Label:         v.Label,
			Content:       v.Content,
			SystemContent: v.SystemContent,
			Temperature:   v.Temperature,
			TopN:          v.TopN,
			MaxTokens:     v.MaxTokens,
			Status:        v.Status,
			CreateTime:    v.Ctime.UnixMilli(),
			UpdateTime:    v.Utime.UnixMilli(),
		}
	}
	return PromptVO{
		ID:            p.ID,
		Name:          p.Name,
		Owner:         p.Owner,
		OwnerType:     p.OwnerType.String(),
		ActiveVersion: p.ActiveVersion,
		Versions:      versions,
		Description:   p.Description,
		CreateTime:    p.Ctime.UnixMilli(),
		UpdateTime:    p.Utime.UnixMilli(),
	}
}

type AddReq struct {
	Name          string  `json:"name"`
	Content       string  `json:"content"`
	Description   string  `json:"description"`
	SystemContent string  `json:"system_content"`
	Temperature   float32 `json:"temperature"`
	TopN          float32 `json:"top_n"`
	MaxTokens     int     `json:"max_tokens"`
}

type DeleteReq struct {
	ID int64 `json:"id"`
}

type DeleteVersionReq struct {
	VersionID int64 `json:"version_id"`
}

type UpdatePromptReq struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type UpdateVersionReq struct {
	VersionID     int64   `json:"version_id,omitempty"`
	Content       string  `json:"content,omitempty"`
	SystemContent string  `json:"system_content"`
	Temperature   float32 `json:"temperature"`
	TopN          float32 `json:"top_n"`
	MaxTokens     int     `json:"max_tokens"`
}

type PublishReq struct {
	VersionID int64  `json:"version_id"`
	Label     string `json:"label"`
}

type ForkReq struct {
	VersionID int64 `json:"version_id"`
}
