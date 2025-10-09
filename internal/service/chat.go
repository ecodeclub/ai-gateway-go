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

	"github.com/gotomicro/ego/core/elog"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

type ChatService struct {
	repo            *repository.ChatRepo
	configRepo      *repository.InvocationConfigRepo
	logger          *elog.Component
	quotaService    *QuotaService
	providerService *ProviderService
}

func NewChatService(
	repo *repository.ChatRepo,
	configRepo *repository.InvocationConfigRepo,
	quotaService *QuotaService,
	provider *ProviderService,
) *ChatService {
	return &ChatService{
		repo:            repo,
		configRepo:      configRepo,
		quotaService:    quotaService,
		providerService: provider,
		logger:          elog.DefaultLogger.With(elog.FieldComponent("service.ChatService")),
	}
}

func (c *ChatService) Save(ctx context.Context, chat domain.ChatV1) (string, error) {
	return c.repo.Save(ctx, chat)
}

func (c *ChatService) List(ctx context.Context, uid int64, limit int64, offset int64) ([]domain.ChatV1, error) {
	return c.repo.GetByUid(ctx, uid, limit, offset)
}

func (c *ChatService) Detail(ctx context.Context, sn string) (domain.ChatV1, error) {
	return c.repo.DetailV1(ctx, sn)
}
