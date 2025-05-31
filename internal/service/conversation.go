// Copyright 2021 ecodeclub
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

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

type ConversationService struct {
	repo *repository.ConversationRepo
}

func NewConversationService(repo *repository.ConversationRepo) *ConversationService {
	return &ConversationService{repo: repo}
}

func (c *ConversationService) Create(ctx context.Context, conversation domain.Conversation) (string, error) {
	return c.repo.Create(ctx, conversation)
}

func (c *ConversationService) List(ctx context.Context, id string) ([]domain.Message, error) {
	return c.repo.GetList(ctx, id)
}
