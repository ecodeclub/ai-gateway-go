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

package service

import (
	"context"
	"math"
	"time"

	"github.com/ecodeclub/ai-gateway-go/errs"
	"github.com/gotomicro/ego/core/elog"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm"
)

type ChatService struct {
	repo            *repository.ChatRepo
	configRepo      *repository.InvocationConfigRepo
	llmHandler      llm.Handler
	logger          *elog.Component
	quotaService    *QuotaService
	providerService *ProviderService
}

func NewChatService(
	repo *repository.ChatRepo,
	configRepo *repository.InvocationConfigRepo,
	handler llm.Handler,
	quotaService *QuotaService,
	provider *ProviderService,
) *ChatService {
	return &ChatService{
		repo:            repo,
		configRepo:      configRepo,
		llmHandler:      handler,
		quotaService:    quotaService,
		providerService: provider,
		logger:          elog.DefaultLogger.With(elog.FieldComponent("service.ChatService")),
	}
}

func (c *ChatService) Save(ctx context.Context, chat domain.Chat) (string, error) {
	return c.repo.Save(ctx, chat)
}

func (c *ChatService) List(ctx context.Context, uid int64, limit int64, offset int64) ([]domain.Chat, error) {
	return c.repo.GetByUid(ctx, uid, limit, offset)
}

func (c *ChatService) Detail(ctx context.Context, sn string) (domain.Chat, error) {
	return c.repo.Detail(ctx, sn)
}

func (c *ChatService) Stream(ctx context.Context, req domain.ChatStreamRequest) (chan domain.StreamEvent, error) {
	ok, err := c.quotaService.HasEnoughQuota(ctx, req.Uid)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errs.ErrAccountOverdue
	}

	configVersion, err := c.configRepo.GetActiveVersionByID(ctx, req.InvocationConfigID)
	if err != nil {
		return nil, err
	}

	model, err := c.providerService.ModelDetail(ctx, configVersion.Model.ID)
	if err != nil {
		return nil, err
	}
	configVersion.Model = model

	err = c.repo.AddMessages(ctx, req.Sn, req.Messages)
	if err != nil {
		return nil, err
	}

	llmEvents, err := c.llmHandler.Stream(ctx, domain.StreamRequest{
		Messages:      req.Messages,
		ConfigVersion: configVersion,
	})
	if err != nil {
		return nil, err
	}
	events := make(chan domain.StreamEvent, 10)
	go c.forward(ctx, model, req, llmEvents, events)
	return events, nil
}

func (c *ChatService) forward(ctx context.Context, model domain.Model, req domain.ChatStreamRequest, in, out chan domain.StreamEvent) {
	content := ""
	reasoningContent := ""
	var (
		inputToken  int64
		outputToken int64
	)
	defer func() {
		saveCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err1 := c.repo.AddMessages(saveCtx, req.Sn, []domain.Message{{
			Content:          content,
			ReasoningContent: reasoningContent,
		}})
		if err1 != nil {
			c.logger.Error("写入数据库失败", elog.FieldErr(err1))
		}
		amount := int64(
			float64(model.InputPrice)*float64(inputToken)/1000 +
				float64(model.OutputPrice)*float64(outputToken)/1000 + 0.5,
		)
		c.deduct(req.Uid, req.Key, amount)
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-in:
			if !ok || evt.Done {
				inputToken += evt.InputToken
				outputToken += evt.OutputToken
				out <- domain.StreamEvent{Done: true}
				return
			}
			reasoningContent += evt.ReasoningContent
			content += evt.Content
			out <- evt
		}
	}
}

func (c *ChatService) deduct(uid int64, key string, amount int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	maxRetry := 3
	for i := 0; i < maxRetry; i++ {
		err := c.quotaService.Deduct(ctx, uid, amount, key)
		if err != nil {
			c.logger.Error("扣减失败",
				elog.FieldErr(err),
				elog.Int64("uid", uid),
				elog.String("key", key),
				elog.Int64("amount", amount),
			)
		}
		time.Sleep(time.Second * time.Duration(math.Pow(2, float64(i))))
	}
}
