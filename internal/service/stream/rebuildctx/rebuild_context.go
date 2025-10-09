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

package rebuildctx

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/gotomicro/ego/core/elog"
	"golang.org/x/sync/errgroup"
)

// RebuildContextHandler 重建整个对话上下文
type RebuildContextHandler struct {
	repo    *repository.ChatRepo
	cfgRepo *repository.InvocationConfigRepo
	priRepo *repository.ProviderRepository
	Next    stream.Handler
	logger  *elog.Component
}

func NewRebuildContextHandler(
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
	if ctx.Initiated {
		return r.Next.Stream(ctx)
	}
	// 即便是执行失败，我们也会把这个设置为 true
	// 不能使用 defer initiated，因为 r.Next.Stream 会在彻底结束之后才返回
	ctx.Initiated = true
	// 构建当前的 Turn，一般来说步骤不会超过 4 个
	steps := make([]*domain.Step, 0, 4)
	// 默认第一个步骤就是调用 LLM，这个假设不要轻易破坏了
	// 后续如果有别的可能，那么就要考虑扩展 invocation_config 与这个地方了
	stepData := &domain.StepLLMData{}
	steps = append(steps, &domain.Step{
		Type: domain.StepTypeLLM,
		Data: stepData,
	})

	var eg errgroup.Group
	eg.Go(func() error {
		chat, err := r.repo.DetailV1(ctx.Ctx, ctx.Chat.Sn)
		ctx.Chat = chat
		return err
	})

	eg.Go(func() error {
		cfg, err := r.cfgRepo.GetActiveVersionByID(ctx.Ctx, ctx.CfgID)
		if err != nil {
			return err
		}
		model, err := r.priRepo.GetModel(ctx.Ctx, cfg.Model.ID)
		cfg.Model = model
		stepData.Cfg = cfg
		return err
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	ctx.Chat.Turns = append(ctx.Chat.Turns, &domain.Turn{
		Vars: map[string]any{
			"Input": ctx.Input.Content,
		},
		UserRun: &domain.UserRun{
			Content: ctx.Input.Content,
			Files:   ctx.Input.Files,
		},
		AssistantRun: &domain.AssistantRun{
			Steps: steps,
		},
	})
	return r.Next.Stream(ctx)
}
