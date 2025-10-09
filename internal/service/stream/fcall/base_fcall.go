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

package fcall

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/gotomicro/ego/core/elog"
)

type BaseFCall struct {
	handler stream.Handler
	cfgRepo *repository.InvocationConfigRepo
	logger  *elog.Component
}

func NewBaseFCall(handler stream.Handler, cfgRepo *repository.InvocationConfigRepo) *BaseFCall {
	return &BaseFCall{
		handler: handler,
		cfgRepo: cfgRepo,
		logger:  elog.DefaultLogger.With(elog.FieldComponentName("BaseFCall")),
	}
}

func (b *BaseFCall) InvokeLLM(ctx *domain.StreamContext, cfgID int64) error {
	b.logger.Debug("收到 invoke_llm 请求", elog.Int64("cfgID", cfgID))
	cfg, err := b.cfgRepo.GetActiveVersionByID(ctx.Ctx, cfgID)
	if err != nil {
		return err
	}
	assistant := ctx.Chat.LastTurn().AssistantRun
	// 插入一个新的 step，并且把数据搞好
	assistant.StartLLMStep(cfg)
	return b.handler.Stream(ctx)
}
