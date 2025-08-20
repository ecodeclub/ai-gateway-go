// Copyright 2025 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package admin

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ekit/slice"
)

type DeleteReq struct {
	ID int64 `json:"id"`
}

type SaveGraphReq struct {
	ID    int64  `json:"id"`
	Steps []Node `json:"steps"`
	Edges []Edge `json:"edges"`
}

type GraphVO struct {
	ID    int64  `json:"id"`
	Nodes []Node `json:"steps"`
	Edges []Edge `json:"edges"`
}

type GetReq struct {
	ID int64 `json:"id"`
}

type Node struct {
	ID       int64  `json:"id,omitempty"`
	GraphID  int64  `json:"graph_id,omitempty"`
	Type     string `json:"type,omitempty"`
	Status   string `json:"status,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}

type Edge struct {
	ID       int64  `json:"id,omitempty"`
	GraphID  int64  `json:"graph_id,omitempty"`
	SourceID int64  `json:"source_id,omitempty"`
	TargetID int64  `json:"target_id,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}

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
