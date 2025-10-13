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

package savedoc

import (
	"encoding/json"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/gotomicro/ego/core/elog"
)

// FCall 保存文档的函数调用
// 我觉得这个功能和 emit_json 有一点重复，可以考虑用这个取代掉 emit_json
// 这个实现的关键点就是会把 Doc 放入到 ctx.Chat.Vars 里面
type FCall struct {
	logger *elog.Component
}

func NewFCall() *FCall {
	return &FCall{
		logger: elog.DefaultLogger.With(elog.FieldComponent("fcall.save_doc"))}
}

func (c *FCall) Name() string {
	return "save_doc"
}

func (c *FCall) Call(ctx *domain.StreamContext, req fcall.Request) (fcall.Response, error) {
	var saveReq Request
	err := json.Unmarshal(req.Args, &saveReq)
	if err != nil {
		return fcall.Response{}, err
	}
	c.logger.Debug("保存变量", elog.String("varName", saveReq.VarName), elog.String("type", saveReq.Type), elog.String("content", saveReq.Content))
	ctx.Chat.Vars[saveReq.VarName] = saveReq.Content
	return fcall.Response{
		NextInvCfgID: saveReq.NextInvCfgID,
	}, err
}

type Request struct {
	VarName      string `json:"varName,omitempty"`
	Type         string `json:"type,omitempty"`
	Content      string `json:"content,omitempty"`
	NextInvCfgID int64  `json:"nextInvCfgId,omitempty"`
}
