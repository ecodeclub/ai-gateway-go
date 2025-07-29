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

package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/repository/sn"
	"golang.org/x/sync/errgroup"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/cache"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

type ChatRepo struct {
	dao   *dao.ChatDAO
	cache *cache.ChatCache
	snGen *sn.Generator
}

func NewChatRepo(d *dao.ChatDAO, c *cache.ChatCache) *ChatRepo {
	return &ChatRepo{dao: d,
		cache: c,
		// 目前没有别的实现，先写死
		snGen: &sn.Generator{},
	}
}

func (repo *ChatRepo) Save(ctx context.Context, chat domain.Chat) (string, error) {
	if chat.Sn == "" {
		chat.Sn = repo.snGen.Generate(chat.Uid)
	}
	if chat.Title == "" {
		chat.Title = fmt.Sprintf("Chat-%s", chat.Sn)
	}
	err := repo.dao.Save(ctx, dao.Chat{
		Sn:    chat.Sn,
		Title: chat.Title,
		Uid:   chat.Uid,
	})
	return chat.Sn, err
}

func (repo *ChatRepo) AddMessages(ctx context.Context, chatSN string, messages []domain.Message) error {
	err := repo.dao.AddMessages(ctx, repo.toDaoMessage(chatSN, messages))
	if err != nil {
		return err
	}
	err = repo.cache.AddMessages(ctx, chatSN, repo.toCacheMessage(messages)...)
	if err != nil {
		elog.Error(fmt.Sprintf("写入redis 失败: %s", chatSN), elog.Any("err", err))
	}
	return nil
}

// GetByUid 根据 uid 获取对话列表
func (repo *ChatRepo) GetByUid(ctx context.Context, uid int64, limit int64, offset int64) ([]domain.Chat, error) {
	chat, err := repo.dao.GetByUid(ctx, uid, limit, offset)
	if err != nil {
		return []domain.Chat{}, err
	}
	return repo.toChats(chat), nil
}

// GetHistoryMessageList 用来获取历史消息列表
func (repo *ChatRepo) GetHistoryMessageList(ctx context.Context, sn string) ([]domain.Message, error) {
	messageCache, err := repo.cache.GetMessages(ctx, sn)
	if err != nil {
		messages, err := repo.dao.GetMessages(ctx, sn)
		if err != nil {
			return []domain.Message{}, err
		}

		domainMessages := repo.toDomainMessage(messages)
		err = repo.cache.AddMessages(ctx, sn, repo.toCacheMessage(domainMessages)...)
		if err != nil {
			elog.Error(fmt.Sprintf("消息写入redis 失败: %s", sn), elog.Any("err", err))
		}
		return repo.toDomainMessage(messages), nil
	}
	return repo.toMessage(messageCache), nil
}

func (repo *ChatRepo) Detail(ctx context.Context, sn string) (domain.Chat, error) {
	var (
		eg       errgroup.Group
		messages []dao.Message
		chat     dao.Chat
	)
	eg.Go(func() error {
		var err error
		chat, err = repo.dao.GetBySN(ctx, sn)
		return err
	})
	eg.Go(func() error {
		var err error
		messages, err = repo.dao.GetMessages(ctx, sn)
		return err
	})

	err := eg.Wait()

	return domain.Chat{
		Sn:       chat.Sn,
		Uid:      chat.Uid,
		Title:    chat.Title,
		Ctime:    time.UnixMilli(chat.Ctime),
		Messages: repo.toDomainMessage(messages),
	}, err
}

func (repo *ChatRepo) toDaoMessage(chatSN string, messages []domain.Message) []dao.Message {
	return slice.Map[domain.Message, dao.Message](messages, func(idx int, src domain.Message) dao.Message {
		return dao.Message{
			ID:            src.ID,
			ChatSN:        chatSN,
			Role:          src.Role,
			Content:       src.Content,
			ReasonContent: src.ReasoningContent,
		}
	})
}

func (repo *ChatRepo) toDomainMessage(messages []dao.Message) []domain.Message {
	return slice.Map[dao.Message, domain.Message](messages, func(idx int, src dao.Message) domain.Message {
		return domain.Message{
			ID:               src.ID,
			Role:             src.Role,
			Content:          src.Content,
			ReasoningContent: src.ReasonContent,
		}
	})
}

func (repo *ChatRepo) toCacheMessage(messages []domain.Message) []cache.Message {
	return slice.Map[domain.Message, cache.Message](messages, func(idx int, src domain.Message) cache.Message {
		return cache.Message{
			Role:          src.Role,
			Content:       src.Content,
			ReasonContent: src.ReasoningContent,
		}
	})
}

func (repo *ChatRepo) toMessage(messages []cache.Message) []domain.Message {
	return slice.Map[cache.Message, domain.Message](messages, func(idx int, src cache.Message) domain.Message {
		return domain.Message{
			Role:             src.Role,
			Content:          src.Content,
			ReasoningContent: src.ReasonContent,
		}
	})
}

func (repo *ChatRepo) toChats(chats []dao.Chat) []domain.Chat {
	return slice.Map(chats, func(idx int, src dao.Chat) domain.Chat {
		return domain.Chat{
			Sn:    src.Sn,
			Title: src.Title,
			Ctime: time.UnixMilli(src.Ctime),
		}
	})
}
