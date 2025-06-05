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

package test

import (
	"context"
	"testing"

	aiv1 "github.com/ecodeclub/ai-gateway-go/api/gen/ai/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/cache"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/test/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yumosx/got/pkg/config"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

type ConversationSuite struct {
	suite.Suite
	db    *gorm.DB
	cache redis.Cmdable
}

func NewConversationSuite() *ConversationSuite {
	return &ConversationSuite{}
}

func TestConversation(t *testing.T) {
	suite.Run(t, NewConversationSuite())
}

func (c *ConversationSuite) SetupTest() {
	dbConfig := config.NewConfig(
		config.WithDBName("ai_gateway_platform"),
		config.WithUserName("root"),
		config.WithPassword("root"),
		config.WithHost("127.0.0.1"),
		config.WithPort("13306"),
	)
	db, err := config.NewDB(dbConfig)
	require.NoError(c.T(), err)

	cacheConfig := config.NewCacheConfig(
		config.WithAddr("localhost:6379"),
	)

	rdb := config.NewCache(cacheConfig)
	c.db = db
	c.cache = rdb
}

func (c *ConversationSuite) TearDownTest() {
	err := c.db.Exec("TRUNCATE TABLE messages").Error
	require.NoError(c.T(), err)
	err = c.db.Exec("TRUNCATE TABLE conversations").Error
	require.NoError(c.T(), err)
}

func (c *ConversationSuite) TestCreate() {
	t := c.T()
	testcases := []struct {
		name   string
		before func()
		after  func()
	}{
		{
			name: "创建对应的 conversation",
			after: func() {
				var conversation dao.Conversation
				err := c.db.Where("id = ?", 1).First(&conversation).Error
				require.NoError(t, err)
				assert.Equal(t, "test", conversation.Title)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			conversationDao := dao.NewConversationDao(c.db)
			conversationCache := cache.NewConversationCache(c.cache)
			repo := repository.NewConversationRepo(conversationDao, conversationCache)
			ctrl := gomock.NewController(t)
			handler := mocks.NewMockHandler(ctrl)
			conversationService := service.NewConversationService(repo, handler)
			server := grpc.NewConversationServer(conversationService)

			res, err := server.Create(context.Background(), &aiv1.Conversation{Title: "test"})
			require.NoError(t, err)
			assert.Equal(t, "1", res.Sn)
			tc.after()
		})
	}
}

func (c *ConversationSuite) TestGetList() {
	t := c.T()
	testcases := []struct {
		name   string
		before func()
		after  func()
	}{
		{
			name: "获取conversation list",
			before: func() {
				err := c.db.Create([]dao.Conversation{
					{Title: "test1", Uid: "123"},
					{Title: "test2", Uid: "123"},
				}).Error
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			conversationDao := dao.NewConversationDao(c.db)
			conversationCache := cache.NewConversationCache(c.cache)
			repo := repository.NewConversationRepo(conversationDao, conversationCache)
			ctrl := gomock.NewController(t)
			handler := mocks.NewMockHandler(ctrl)
			conversationService := service.NewConversationService(repo, handler)
			server := grpc.NewConversationServer(conversationService)
			res, err := server.List(context.Background(), &aiv1.ListReq{Uid: "123", Offset: 0, Limit: 2})
			require.NoError(t, err)
			assert.Equal(t, len(res.Conversations), 2)
		})
	}
}

func (c *ConversationSuite) TestChat() {
	t := c.T()
	testcases := []struct {
		name   string
		before func(handler *mocks.MockHandler)
		after  func()
	}{
		{
			name: "与大模型chat",
			before: func(handler *mocks.MockHandler) {
				err := c.db.Create(&dao.Conversation{Title: "test1", Uid: "123"}).Error
				require.NoError(t, err)
				resp := domain.ChatResponse{Response: domain.Message{Content: "event1"}}
				handler.EXPECT().Handle(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			after: func() {
				//以数据库的数据为准
				var message []dao.Message
				err := c.db.Find(&message).Where("cid = ?", 1).Error
				require.NoError(t, err)
				assert.Equal(t, 3, len(message))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			conversationDao := dao.NewConversationDao(c.db)
			conversationCache := cache.NewConversationCache(c.cache)
			repo := repository.NewConversationRepo(conversationDao, conversationCache)
			ctrl := gomock.NewController(t)
			handler := mocks.NewMockHandler(ctrl)
			conversationService := service.NewConversationService(repo, handler)
			server := grpc.NewConversationServer(conversationService)

			tc.before(handler)
			chat, err := server.Chat(context.Background(), &aiv1.Conversation{Sn: "1", Uid: "123",
				Message: []*aiv1.Message{
					{Content: "content1"},
					{Content: "content2"},
				}})
			require.NoError(t, err)
			assert.Equal(t, chat.Sn, "1")
			assert.Equal(t, chat.Response.Content, "event1")
			tc.after()
		})
	}
}

func (c *ConversationSuite) TestStream() {
	t := c.T()
	testcases := []struct {
		name   string
		before func(handler *mocks.MockHandler)
		after  func()
		want   []domain.StreamEvent
	}{
		{
			name: "流式传输",
			before: func(handler *mocks.MockHandler) {
				streamChan := make(chan domain.StreamEvent, 2)
				streamChan <- domain.StreamEvent{Content: "event1", ReasoningContent: "reason1"}
				streamChan <- domain.StreamEvent{Content: "event2", ReasoningContent: "reason1"}
				close(streamChan)
				handler.EXPECT().StreamHandle(gomock.Any(), gomock.Any()).Return(streamChan, nil)
				err := c.db.Create(&dao.Conversation{Title: "test1", Uid: "123"}).Error
				require.NoError(t, err)
			},
			after: func() {
				//以数据库的数据为准
				var message []dao.Message
				err := c.db.Find(&message).Where("cid = ?", 1).Error
				require.NoError(t, err)
				assert.Equal(t, 3, len(message))
			},
			want: []domain.StreamEvent{{Content: "event1", ReasoningContent: "reason1"}, {Content: "event2", ReasoningContent: "reason2"}},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			conversationDao := dao.NewConversationDao(c.db)
			conversationCache := cache.NewConversationCache(c.cache)
			repo := repository.NewConversationRepo(conversationDao, conversationCache)
			ctrl := gomock.NewController(t)
			handler := mocks.NewMockHandler(ctrl)
			conversationService := service.NewConversationService(repo, handler)
			server := grpc.NewConversationServer(conversationService)
			tc.before(handler)
			mockStream := &mocks.MockStreamServer{Ctx: context.Background()}
			err := server.Stream(&aiv1.Conversation{Sn: "1", Uid: "123",
				Message: []*aiv1.Message{
					{Content: "content1"},
					{Content: "content2"},
				}}, mockStream)
			require.NoError(t, err)
			for i, event := range tc.want {
				assert.Equal(t, event.Content, mockStream.Events[i].Content)
			}
			tc.after()
		})
	}
}
