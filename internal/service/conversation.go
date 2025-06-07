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
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm"
	"github.com/gotomicro/ego/core/elog"
)

type ConversationService struct {
	repo   *repository.ConversationRepo
	handle llm.Handler
}

func NewConversationService(repo *repository.ConversationRepo, handler llm.Handler) *ConversationService {
	return &ConversationService{repo: repo, handle: handler}
}

func (c *ConversationService) Create(ctx context.Context, conversation domain.Conversation) (string, error) {
	return c.repo.Create(ctx, conversation)
}

func (c *ConversationService) List(ctx context.Context, uid string, limit int64, offset int64) ([]domain.Conversation, error) {
	return c.repo.GetByUid(ctx, uid, limit, offset)
}

func (c *ConversationService) Chat(ctx context.Context, sn string, messages []domain.Message) (domain.ChatResponse, error) {
	err := c.repo.AddMessages(ctx, sn, messages)
	if err != nil {
		return domain.ChatResponse{}, err
	}

	messageList, err := c.repo.GetMessageList(ctx, sn, 20, 0)
	if err != nil {
		return domain.ChatResponse{}, err
	}

	response, err := c.handle.Handle(ctx, messageList)
	if err != nil {
		return domain.ChatResponse{}, err
	}

	resp := domain.ChatResponse{Sn: sn, Response: response.Response}

	// 将返回结果写入repo
	err = c.repo.AddMessages(ctx, sn, []domain.Message{response.Response})
	if err != nil {
		return domain.ChatResponse{}, err
	}

	return resp, nil
}

func (c *ConversationService) Stream(ctx context.Context, sn string, messages []domain.Message) (chan domain.StreamEvent, error) {
	ch := make(chan domain.StreamEvent, 10)

	err := c.repo.AddMessages(ctx, sn, messages)
	if err != nil {
		return ch, err
	}

	cs, err := c.repo.GetMessageList(ctx, sn, 20, 0)
	if err != nil {
		return ch, err
	}

	event, err := c.handle.StreamHandle(ctx, cs)
	if err != nil {
		return ch, err
	}

	go func() {
		var message domain.Message

		for {
			select {
			case <-ctx.Done():
				return
			case value, ok := <-event:
				if !ok || value.Done {
					err = c.repo.AddMessages(ctx, sn, []domain.Message{message})
					if err != nil {
						elog.Error("写入数据库失败", elog.FieldErr(err))
					}
					ch <- domain.StreamEvent{Done: true}
					return
				}

				message.ReasoningContent += value.ReasoningContent
				message.ReasoningContent += value.Content
				ch <- value
			}
		}
	}()
	return ch, err
}
