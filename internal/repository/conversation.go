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
	res, err := repo.dao.Create(ctx, dao.Conversation{Title: conversation.Title, Uid: conversation.Uid})
	if err != nil {
		return "", err
	}

	return strconv.Itoa(int(res.ID)), nil
}

func (repo *ConversationRepo) CreateMessages(ctx context.Context, conversation domain.Conversation) error {
	cid, _ := strconv.Atoi(conversation.Sn)

	err := repo.dao.CreateMessages(ctx, repo.toDaoMessage(int64(cid), conversation.Messages))
	if err != nil {
		return err
	}

	err = repo.cache.AddMessages(ctx, conversation.Sn, conversation.Uid, repo.toCacheMessage(conversation.Messages))
	if err != nil {
		elog.Error(fmt.Sprintf("用户 %s 写入redis 失败: %s", conversation.Uid, conversation.Sn), elog.Any("err", err))
	}
	return nil
}

func (repo *ConversationRepo) GetByUid(ctx context.Context, uid string, limit int64, offset int64) ([]domain.Conversation, error) {
	conversation, err := repo.dao.GetByUid(ctx, uid, limit, offset)
	if err != nil {
		return []domain.Conversation{}, err
	}
	return repo.toConversation(conversation), nil
}

// GetById 用来每次对话时候获取对话的消息
func (repo *ConversationRepo) GetById(ctx context.Context, id int64, limit int64, offset int64) (domain.Conversation, error) {
	conversation, err := repo.dao.GetById(ctx, id)
	if err != nil {
		return domain.Conversation{}, err
	}

	cid := strconv.Itoa(int(conversation.ID))
	list, err := repo.getMessageList(ctx, strconv.Itoa(int(id)), cid, limit, offset)
	if err != nil {
		return domain.Conversation{}, err
	}

	return domain.Conversation{Sn: cid, Title: conversation.Title, Messages: list}, nil
}

func (repo *ConversationRepo) getMessageList(ctx context.Context, cid string, uid string, limit int64, offset int64) ([]domain.Message, error) {
	messageCache, err := repo.cache.GetMessage(ctx, cid, uid, limit, offset)
	if err != nil {
		ID, _ := strconv.Atoi(cid)
		messages, err := repo.dao.GetMessages(ctx, int64(ID), limit, offset)
		if err != nil {
			return []domain.Message{}, err
		}

		domainMessages := repo.toDomainMessage(messages)
		err = repo.cache.AddMessages(ctx, cid, uid, repo.toCacheMessage(domainMessages))
		if err != nil {
			elog.Error(fmt.Sprintf("用户 %s 的消息写入redis 失败: %s", uid, cid), elog.Any("err", err))
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
			CID:              src.CID,
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

func (repo *ConversationRepo) toConversation(conversations []dao.Conversation) []domain.Conversation {
	return slice.Map(conversations, func(idx int, src dao.Conversation) domain.Conversation {
		return domain.Conversation{
			Sn:    strconv.Itoa(int(src.ID)),
			Title: src.Title,
		}
	})
}
