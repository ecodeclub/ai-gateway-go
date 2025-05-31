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

package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/cache"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

type ConversationRepo struct {
	dao   *dao.ConversationDao
	cache *cache.ConversationCache
}

func NewConversationRepo(d *dao.ConversationDao, c *cache.ConversationCache) *ConversationRepo {
	return &ConversationRepo{dao: d, cache: c}
}

func (repo *ConversationRepo) Create(ctx context.Context, conversation domain.Conversation) (string, error) {
	res, err := repo.dao.Create(ctx, dao.Conversation{})
	if err != nil {
		return "", err
	}

	if len(conversation.Messages) != 0 {
		err := repo.dao.CreateMsgs(ctx, repo.toDaoMessage(res.ID, conversation.Messages))
		if err != nil {
			return "", err
		}

		// 写到 redis 中
		err = repo.cache.AddMessages(ctx, strconv.Itoa(int(res.ID)), repo.toCacheMessage(conversation.Messages))
		if err != nil {
			elog.Error(fmt.Sprintf("写入redis 失败: %d", res.ID), elog.Any("err", err))
		}
	}
	return strconv.Itoa(int(res.ID)), nil
}

func (repo *ConversationRepo) GetList(ctx context.Context, id string) ([]domain.Message, error) {
	ID, _ := strconv.Atoi(id)

	// 首先从redis 中去查找对应消息
	messageCache, err := repo.cache.GetMessage(ctx, id)
	if err != nil {
		messages, err := repo.dao.GetMessages(ctx, int64(ID))
		if err != nil {
			return []domain.Message{}, err
		}

		domainMessages := repo.toDomainMessage(messages)
		err = repo.cache.AddMessages(ctx, id, repo.toCacheMessage(domainMessages))
		if err != nil {
			elog.Error(fmt.Sprintf("写入redis 失败: %s", id), elog.Any("err", err))
		}
		return repo.toDomainMessage(messages), nil
	}
	return repo.toMessage(messageCache), nil
}

func (repo *ConversationRepo) toDaoMessage(id int64, messages []domain.Message) []dao.Message {
	return slice.Map[domain.Message, dao.Message](messages, func(idx int, src domain.Message) dao.Message {
		return dao.Message{
			ID:            src.ID,
			CID:           id,
			Role:          src.Role,
			Content:       src.Content,
			ReasonContent: src.ReasoningContent,
		}
	})
}

func (repo *ConversationRepo) toDomainMessage(messages []dao.Message) []domain.Message {
	return slice.Map[dao.Message, domain.Message](messages, func(idx int, src dao.Message) domain.Message {
		return domain.Message{
			ID:               src.ID,
			Role:             src.Role,
			Content:          src.Content,
			ReasoningContent: src.ReasonContent,
		}
	})
}

func (repo *ConversationRepo) toCacheMessage(messages []domain.Message) []cache.Message {
	return slice.Map[domain.Message, cache.Message](messages, func(idx int, src domain.Message) cache.Message {
		return cache.Message{
			Role:          src.Role,
			Content:       src.Content,
			ReasonContent: src.ReasoningContent,
		}
	})
}

func (repo *ConversationRepo) toMessage(messages []cache.Message) []domain.Message {
	return slice.Map[cache.Message, domain.Message](messages, func(idx int, src cache.Message) domain.Message {
		return domain.Message{
			Role:             src.Role,
			Content:          src.Content,
			ReasoningContent: src.ReasonContent,
		}
	})
}
