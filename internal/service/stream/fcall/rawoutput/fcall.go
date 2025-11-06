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

package rawoutput

import (
	"encoding/json"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/gotomicro/ego/core/elog"
)

// FCall 保存文档的函数调用
type FCall struct {
	logger *elog.Component
}

func NewFCall() *FCall {
	return &FCall{
		logger: elog.DefaultLogger.With(elog.FieldComponent("fcall.raw_output"))}
}

func (c *FCall) Name() string {
	return "raw_output"
}

func (c *FCall) Call(ctx *domain.StreamContext, req fcall.Request) (fcall.Response, error) {
	var rawReq Request
	err := json.Unmarshal(req.Args, &rawReq)
	if err != nil {
		return fcall.Response{}, err
	}
	c.logger.Debug("收到 RawOutput", elog.Any("rawReq", rawReq))
	const varName = "RawOutput"
	ctx.Chat.Vars[varName] = rawReq.Content
	return fcall.Response{
		Content:   rawReq.Content,
		NextState: rawReq.State,
	}, err
}

type Request struct {
	Content string `json:"content"`
	State   string `json:"state"`
}
