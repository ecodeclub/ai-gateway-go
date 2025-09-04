package test

import (
	"context"
	"testing"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/repository/cache"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service/mocks"
	testioc "github.com/ecodeclub/ai-gateway-go/internal/test/ioc"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type ChatSuite struct {
	suite.Suite
	*testioc.TestApp
	dao     *dao.ChatDAO
	cache   *cache.ChatCache
	handler *mocks.MockHandler
}

func TestChat(t *testing.T) {
	t.Skip()
	suite.Run(t, &ChatSuite{})
}

func (c *ChatSuite) SetupSuite() {
	ctrl := gomock.NewController(c.T())
	handler := mocks.NewMockHandler(ctrl)
	c.handler = handler
	app := testioc.InitApp(testioc.TestOnly{LLM: handler})
	c.TestApp = app
	c.dao = dao.NewChatDAO(c.DB)
	c.cache = cache.NewChatCache(c.Rdb)
}

func (c *ChatSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := c.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE chats").Error
	if err != nil {
		c.T().Log(err)
	}
	err = c.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE messages").Error
	if err != nil {
		c.T().Log(err)
	}

	c.TestApp.Rdb.FlushDB(ctx)
}
