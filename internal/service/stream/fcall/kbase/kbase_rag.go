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

package kbase

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/ecodeclub/ekit/net/httpx"
	"github.com/gotomicro/ego/core/elog"
)

type RAG struct {
	client *http.Client
	url    string
	logger *elog.Component
}

// NewKBaseRAG 创建 RAG 实例
// url 是知识库的查询地址
func NewKBaseRAG(url string) *RAG {
	return &RAG{
		logger: elog.DefaultLogger.With(elog.FieldComponent("KBaseRAG")),
		client: http.DefaultClient,
		url:    url,
	}
}

func (k *RAG) Name() string {
	return "kbase_rag"
}

func (k *RAG) Call(ctx *domain.StreamContext, req fcall.Request) (fcall.Response, error) {
	var ragReq Request
	err := json.Unmarshal(req.Args, &ragReq)
	if err != nil {
		return fcall.Response{}, fmt.Errorf("kbase_rag call Unmarshal err: %v", err)
	}

	// 规避 JSON 转义问题
	k.logger.Debug("ragReq.Query："+string(ragReq.Query), elog.String("varName", ragReq.VarName))

	response := httpx.NewRequest(ctx.Ctx, http.MethodPost, k.url).
		Client(k.client).JSONBody(ragReq.Query).Do()

	//if response.Err() != nil {
	//	return fcall.Response{}, fmt.Errorf("执行知识库查询失败, err: %v", response.Err())
	//}

	if response.StatusCode != http.StatusOK {
		return fcall.Response{}, fmt.Errorf("执行知识库查询失败, status code: %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fcall.Response{}, err
	}
	k.logger.Debug("RAG 响应", elog.String("body", string(body)))
	ctx.Chat.Vars[ragReq.VarName] = string(body)
	return fcall.Response{
		NextInvCfgID: ragReq.NextInvCfgID,
	}, err
}

type Request struct {
	// 参数名字，也就是放入到 Turn.AssistantRun.Vars 中的 key
	VarName string `json:"varName"`
	// 如果指定了 NextInvCfgID，则在查询完成后，继续调用 LLM
	NextInvCfgID int64 `json:"nextInvCfgID"`
	// 参考 kbase 中 search 接口的入参
	Query json.RawMessage `json:"query"`
}
