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

package loadcfg

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/gotomicro/ego/core/elog"
)

// RebuildContextHandler 重建整个对话上下文
type RebuildContextHandler struct {
	repo    *repository.ChatRepo
	cfgRepo *repository.InvocationConfigRepo
	priRepo *repository.ProviderRepository
	Next    stream.Handler
	logger  *elog.Component
}

func NewLoadConfigHandler(
	repo *repository.ChatRepo,
	cfgRepo *repository.InvocationConfigRepo,
	priRepo *repository.ProviderRepository,
) *RebuildContextHandler {
	return &RebuildContextHandler{
		repo:    repo,
		cfgRepo: cfgRepo,
		priRepo: priRepo,
		logger:  elog.DefaultLogger.With(elog.FieldComponent("handler.rebuild_ctx")),
	}
}

// Stream 是否移动到 service 会更好？
func (r *RebuildContextHandler) Stream(ctx *domain.StreamContext) error {
	stepData := ctx.Chat.LastTurn().AssistantRun.LastStep().LLMData()
	cfg, err := r.cfgRepo.GetActiveVersionByID(ctx.Ctx, stepData.CfgID)
	if err != nil {
		return err
	}
	model, err := r.priRepo.GetModel(ctx.Ctx, cfg.Model.ID)
	cfg.Model = model
	stepData.Cfg = cfg
	if err != nil {
		return err
	}
	return r.Next.Stream(ctx)
}
