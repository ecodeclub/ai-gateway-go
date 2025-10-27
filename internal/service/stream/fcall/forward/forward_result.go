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

package forward

import (
	"encoding/json"
	"fmt"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/gotomicro/ego/core/elog"
)

type Request struct {
	// 参数名字，也就是放入到 Turn.AssistantRun.Vars 中的 key
	VarName string `json:"varName"`
	// 如果指定了 NextInvCfgID，则在查询完成后，继续调用 LLM
	NextInvCfgID int64           `json:"nextInvCfgID"`
	Result       json.RawMessage `json:"result"`
}

type Result struct {
	logger *elog.Component
}

// NewResult 创建 Result 实例 用于转发JSON
func NewResult() *Result {
	return &Result{
		logger: elog.DefaultLogger.With(elog.FieldComponent("forward.Result")),
	}
}

func (r *Result) Name() string {
	return "forward_result"
}

func (r *Result) Call(ctx *domain.StreamContext, req fcall.Request) (fcall.Response, error) {
	var resReq Request
	err := json.Unmarshal(req.Args, &resReq)
	if err != nil {
		return fcall.Response{}, fmt.Errorf("%s call Unmarshal err: %v", r.Name(), err)
	}

	// 规避 JSON 双重转义问题（百炼兼容性处理）
	// 检测 result 是否被错误地序列化为字符串，如果是则解开一层
	var resultStr string
	if err := json.Unmarshal(resReq.Result, &resultStr); err == nil {
		// resReq.Result 是一个 JSON 字符串（双重转义），需要解开
		r.logger.Debug("检测到双重转义，自动解开",
			elog.String("原始值", string(resReq.Result)),
			elog.String("解开后", resultStr))
		resReq.Result = []byte(resultStr)
	}

	r.logger.Debug("resReq.Result："+string(resReq.Result), elog.String("varName", resReq.VarName))

	err = ctx.Sender.Send(domain.StreamEventV1{
		Delta: &domain.Delta{
			Content: string(resReq.Result),
		},
	})
	if err != nil {
		return fcall.Response{}, fmt.Errorf("执行转发失败, err: %v", err)
	}

	ctx.Chat.Vars[resReq.VarName] = string(resReq.Result)

	return fcall.Response{
		Content:      "数据已发送给前端，等待用户下一步操作",
		NextInvCfgID: resReq.NextInvCfgID,
	}, err
}
