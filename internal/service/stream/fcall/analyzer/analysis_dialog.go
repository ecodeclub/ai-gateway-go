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
	base   *fcall.BaseFCall
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

	ctx.Chat.Digest = &domain.Digest{
		Summary: anaReq.Summary,
	}
	a.logger.Debug("生成摘要", elog.String("summary", anaReq.Summary))

	if anaReq.Title != "" {
		go func() {
			// 输出一个 title 变更
			err1 := ctx.Sender.Send(domain.StreamEventV1{
				StepUpdate: &domain.StepUpdate{
					Title: anaReq.Title,
					// summary 一般是不需要放的，除非是你在 DEBUG 环境，你想看看 summary 对不对
				},
			})
			if err1 != nil {
				a.logger.Error("发送 title 变更失败", elog.String("title", anaReq.Title), elog.FieldErr(err1))
			}
		}()
	}

	if anaReq.NextInvCfgID > 0 {
		err = a.base.InvokeLLM(ctx, anaReq.NextInvCfgID)
	}
	return fcall.Response{}, err
}

func NewAnalysisDialogFCall(base *fcall.BaseFCall) *AnalysisDialogFCall {
	return &AnalysisDialogFCall{
		base:   base,
		logger: elog.DefaultLogger.With(elog.FieldComponent("AnalysisDialogFCall"))}
}

type AnalysisDialogRequest struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	// 识别出来用户想要干什么
	Intent       string `json:"intent"`
	NextInvCfgID int64  `json:"nextInvCfgID"`
}
