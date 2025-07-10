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
	"github.com/gotomicro/ego/core/elog"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm"
)

type ChatService struct {
	repo   *repository.ChatRepo
	handle llm.Handler
	logger *elog.Component
}

func NewChatService(repo *repository.ChatRepo, handler llm.Handler) *ChatService {
	return &ChatService{repo: repo,
		handle: handler,
		logger: elog.DefaultLogger.With(elog.String("component", "ChatService"))}
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

func (c *ChatService) Stream(ctx context.Context, sn string, messages []domain.Message) (chan domain.StreamEvent, error) {
	ch := make(chan domain.StreamEvent, 10)

	cs, err := c.repo.GetHistoryMessageList(ctx, sn)
	if err != nil {
		return ch, err
	}

	err = c.repo.AddMessages(ctx, sn, messages)
	if err != nil {
		return ch, err
	}

	cs = append(cs, messages...)

	event, err := c.handle.StreamHandle(ctx, cs)
	if err != nil {
		return ch, err
	}

	go func() {
		conent := ""
		reasoningContent := ""
		defer func() {
			saveCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err1 := c.repo.AddMessages(saveCtx, sn, []domain.Message{{
				Content:          conent,
				ReasoningContent: reasoningContent,
			}})
			if err1 != nil {
				c.logger.Error("写入数据库失败", elog.FieldErr(err))
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case value, ok := <-event:
				if !ok || value.Done {
					ch <- domain.StreamEvent{Done: true}
					return
				}
				reasoningContent += value.ReasoningContent
				conent += value.Content
				ch <- value
			}
		}
	}()
	return ch, err
}
