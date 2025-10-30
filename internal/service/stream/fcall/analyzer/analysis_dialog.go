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

package analyzer

import (
	"encoding/json"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/gotomicro/ego/core/elog"
)

type AnalysisDialogFCall struct {
	logger *elog.Component
}

func (a *AnalysisDialogFCall) Name() string {
	return "analysis_dialog"
}

func (a *AnalysisDialogFCall) Call(ctx *domain.StreamContext, req fcall.Request) (fcall.Response, error) {
	var anaReq AnalysisDialogRequest
	err := json.Unmarshal(req.Args, &anaReq)
	if err != nil {
		return fcall.Response{}, err
	}

	if anaReq.Title != "" {
		go func() {
			// 输出一个 title 变更
			err1 := ctx.Sender.Send(domain.StreamEvent{
				StepUpdate: &domain.StepUpdate{
					Title: anaReq.Title,
				},
			})
			if err1 != nil {
				a.logger.Error("发送 title 变更失败", elog.String("title", anaReq.Title), elog.FieldErr(err1))
			}
		}()
	}

	return fcall.Response{
		NextState: anaReq.State,
	}, err
}

func NewAnalysisDialogFCall() *AnalysisDialogFCall {
	return &AnalysisDialogFCall{
		logger: elog.DefaultLogger.With(elog.FieldComponent("AnalysisDialogFCall"))}
}

type AnalysisDialogRequest struct {
	Title string `json:"title"`
	// 识别出来当前多轮对话已经到了什么状态
	State string `json:"state"`
}
