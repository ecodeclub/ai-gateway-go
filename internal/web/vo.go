package web

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ekit/slice"
)

type GetVO struct {
	Name        string `json:"name"`
	Owner       int64  `json:"owner"`
	OwnerType   string `json:"owner_type"`
	Content     string `json:"content"`
	Description string `json:"description"`
	CreateTime  int64  `json:"create_time"`
	UpdateTime  int64  `json:"update_time"`
}

func newGetVO(p domain.Prompt) GetVO {
	return GetVO{
		Name:        p.Name,
		Owner:       p.Owner,
		OwnerType:   p.OwnerType.String(),
		Content:     p.Content,
		Description: p.Description,
		CreateTime:  p.Ctime.UnixMilli(),
		UpdateTime:  p.Utime.UnixMilli(),
	}
}

type AddReq struct {
	Name        string `json:"name"`
	Content     string `json:"content"`
	Description string `json:"description"`
}

type UpdateReq struct {
	Name        string `json:"name,omitempty"`
	Content     string `json:"content,omitempty"`
	Description string `json:"description,omitempty"`
}

type SavePlanReq struct {
	ID    int64  `json:"id"`
	Steps []Step `json:"steps"`
}

type GetPlanVO struct {
	ID    int64  `json:"id"`
	Steps []Step `json:"steps"`
}

type Step struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Metadata string `json:"metadata"`
}

func newGetNodeVO(plan domain.Plan) GetPlanVO {
	var vo GetPlanVO
	vo.ID = plan.ID
	vo.Steps = slice.Map[domain.Step, Step](plan.Steps, func(idx int, src domain.Step) Step {
		m, _ := src.Metadata.AsString()
		return Step{ID: src.ID, Type: src.Type, Status: src.Status, Metadata: m}
	})

	return vo
}
