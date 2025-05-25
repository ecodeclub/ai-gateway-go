package web

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ekit/slice"
)

// Package web 定义了 Web 层的处理器和路由
// vo.go 定义了值对象（Value Object）用于在不同层之间传输数据

// PromptVO 定义提示信息的值对象
// 用于在不同层之间传输提示数据，包括基本信息和多个版本信息
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

// PromptVersionVO 定义提示版本的值对象
// 用于在不同层之间传输提示版本的具体内容
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

// newPromptVO 将领域模型转换为提示值对象
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

// AddReq 定义添加提示的请求参数结构体
type AddReq struct {
	Name          string  `json:"name"`
	Content       string  `json:"content"`
	Description   string  `json:"description"`
	SystemContent string  `json:"system_content"`
	Temperature   float32 `json:"temperature"`
	TopN          float32 `json:"top_n"`
	MaxTokens     int     `json:"max_tokens"`
}

// DeleteReq 定义删除提示的请求参数结构体
type DeleteReq struct {
	ID int64 `json:"id"`
}

// DeleteVersionReq 定义删除提示版本的请求参数结构体
type DeleteVersionReq struct {
	VersionID int64 `json:"version_id"`
}

// UpdatePromptReq 定义更新提示基本信息的请求参数结构体
type UpdatePromptReq struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// SaveGraphReq 定义保存图信息的请求参数结构体
type SaveGraphReq struct {
	ID    int64  `json:"id"`
	Steps []Node `json:"steps"`
	Edges []Edge `json:"edges"`
}

// GraphVO 定义图信息的值对象
type GraphVO struct {
	ID    int64  `json:"id"`
	Nodes []Node `json:"steps"`
	Edges []Edge `json:"edges"`
}

// GetReq 定义获取图或提示信息的请求参数结构体
type GetReq struct {
	ID int64 `json:"id"`
}

// Node 定义节点信息的值对象
type Node struct {
	ID       int64  `json:"id,omitempty"`
	GraphID  int64  `json:"graph_id,omitempty"`
	Type     string `json:"type,omitempty"`
	Status   string `json:"status,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}

// Edge 定义边信息的值对象
type Edge struct {
	ID       int64  `json:"id,omitempty"`
	GraphID  int64  `json:"graph_id,omitempty"`
	SourceID int64  `json:"source_id,omitempty"`
	TargetID int64  `json:"target_id,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}

// newGetNodeVO 将领域模型转换为图值对象
func newGetNodeVO(plan domain.Graph) GraphVO {
	var vo GraphVO
	vo.ID = plan.ID
	vo.Nodes = slice.Map[domain.Node, Node](plan.Steps, func(idx int, src domain.Node) Node {
		m, _ := src.Metadata.AsString()
		return Node{ID: src.ID, Type: src.Type, Status: src.Status, Metadata: m, GraphID: src.GraphID}
	})
	vo.Edges = slice.Map[domain.Edge, Edge](plan.Edges, func(idx int, src domain.Edge) Edge {
		m, _ := src.Metadata.AsString()
		return Edge{ID: src.ID, TargetID: src.TargetID, SourceID: src.SourceID, Metadata: m, GraphID: src.GraphID}
	})
	return vo
}

// UpdateVersionReq 定义更新提示版本信息的请求参数结构体
type UpdateVersionReq struct {
	VersionID     int64   `json:"version_id,omitempty"`
	Content       string  `json:"content,omitempty"`
	SystemContent string  `json:"system_content"`
	Temperature   float32 `json:"temperature"`
	TopN          float32 `json:"top_n"`
	MaxTokens     int     `json:"max_tokens"`
}

// PublishReq 定义发布特定提示版本的请求参数结构体
type PublishReq struct {
	VersionID int64  `json:"version_id"`
	Label     string `json:"label"`
}

// ForkReq 定义复制（fork）提示版本的请求参数结构体
type ForkReq struct {
	VersionID int64 `json:"version_id"`
}
