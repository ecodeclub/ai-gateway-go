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

package store

import (
	"context"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/gotomicro/ego/core/elog"
	"github.com/valyala/bytebufferpool"
	"golang.org/x/sync/errgroup"
)

// Handler 存储数据
// 目前来说，就是存储展示给用户的文本信息
type Handler struct {
	repo   *repository.ChatRepo
	pool   *bytebufferpool.Pool
	Next   stream.Handler
	logger *elog.Component
}

func NewHandler(repo *repository.ChatRepo) *Handler {
	return &Handler{repo: repo,
		pool:   &bytebufferpool.Pool{},
		logger: elog.DefaultLogger.With(elog.FieldComponentName("store.Handler"))}
}

func (h *Handler) Stream(ctx *domain.StreamContext) (stream.Response, error) {
	buffer := h.pool.Get()
	sender := &collectWriter{
		buffer: buffer,
		sender: ctx.Sender, // grpc stream server 的封装
		logger: h.logger.With(elog.FieldComponentName("store.Handler.collectWriter")),
	}
	ctx.Sender = sender
	defer func() {
		storeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		turn := ctx.Chat.CurrentTurn()
		turn.AssistantRun.Content = buffer.String()
		h.pool.Put(buffer)
		var eg errgroup.Group
		eg.Go(func() error {
			// 存储到数据库中
			tid, err := h.repo.SaveTurn(storeCtx, ctx.Chat.Sn, turn)
			turn.ID = tid
			return err
		})
		eg.Go(func() error {
			var err error
			// 之前没有设置标题，现在设置一下
			if sender.title != "" && ctx.Chat.Title == "" {
				ctx.Chat.Title = sender.title
			}
			_, err = h.repo.Save(storeCtx, ctx.Chat)
			return err
		})
		err := eg.Wait()
		if err != nil {
			h.logger.Error("保存该轮次信息失败", elog.FieldErr(err))
		}
	}()
	return h.Next.Stream(ctx)
}

type collectWriter struct {
	buffer *bytebufferpool.ByteBuffer
	sender domain.StreamEventSender
	logger *elog.Component

	title string
}

// Send 将数据收集下来之后，直接转发
func (c *collectWriter) Send(evt domain.StreamEvent) error {
	switch {
	case evt.Delta != nil:
		_, err := c.buffer.WriteString(evt.Delta.Content)
		if err != nil {
			c.logger.Error("收集大模型 stream 接口返回的数据失败", elog.FieldErr(err))
		}
	case evt.StepUpdate != nil:
		// 更新标题
		c.title = evt.StepUpdate.Title
		c.logger.Debug("收到修改 title 事件", elog.String("title", c.title))
	}
	return c.sender.Send(evt)
}
