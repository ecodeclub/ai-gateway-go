package e2e

import (
	"context"
	promptv1 "github.com/ecodeclub/ai-gateway-go/api/gen/prompt/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
	"time"
)

type PromptTestSuite struct {
	suite.Suite
	client promptv1.PromptServiceClient
	db     *gorm.DB
}

func TestPrompt(t *testing.T) {
	suite.Run(t, new(PromptTestSuite))
}

func (s *PromptTestSuite) SetupSuite() {
	cc, err := grpc.NewClient("localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(s.T(), err)
	s.client = promptv1.NewPromptServiceClient(cc)
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:13316)/ai_gateway?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s"))
	require.NoError(s.T(), err)
	s.db = db
}

func (s *PromptTestSuite) TearDownTest() {
	err := s.db.Exec("TRUNCATE TABLE prompts").Error
	require.NoError(s.T(), err)
}

func (s *PromptTestSuite) TestAdd() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, err := s.client.Add(ctx, &promptv1.AddRequest{
		Name:        "test",
		Biz:         "test",
		Pattern:     "test",
		Description: "test",
	})
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Res)

	var got dao.Prompt
	err = s.db.Where("id = ?", 1).First(&got).Error
	require.NoError(s.T(), err)
	require.Equal(s.T(), "test", got.Name)
	require.Equal(s.T(), "test", got.Biz)
	require.Equal(s.T(), "test", got.Pattern)
	require.Equal(s.T(), "test", got.Description)
	require.Equal(s.T(), uint8(1), got.Status)
	require.True(s.T(), got.Ctime > 0)
	require.True(s.T(), got.Ctime > 0)
}

func (s *PromptTestSuite) TestGet() {
	s.TestAdd()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	resp, err := s.client.Get(ctx, &promptv1.GetRequest{
		Id: 1,
	})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
	require.Equal(s.T(), "test", resp.Name)
	require.Equal(s.T(), "test", resp.Biz)
	require.Equal(s.T(), "test", resp.Pattern)
	require.Equal(s.T(), "test", resp.Description)
	require.True(s.T(), resp.CreateTime > 0)
	require.True(s.T(), resp.UpdateTime > 0)
}

func (s *PromptTestSuite) TestUpdate() {
	s.TestAdd()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	resp, err := s.client.Update(ctx, &promptv1.UpdateRequest{
		Id:   1,
		Name: "aaa",
	})
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Res)

	var got dao.Prompt
	err = s.db.Where("id = ?", 1).First(&got).Error
	require.NoError(s.T(), err)
	require.Equal(s.T(), "aaa", got.Name)
	require.Equal(s.T(), "test", got.Biz)
	require.Equal(s.T(), "test", got.Pattern)
	require.Equal(s.T(), "test", got.Description)
	require.Equal(s.T(), uint8(1), got.Status)
	require.True(s.T(), got.Utime > got.Ctime)

}

func (s *PromptTestSuite) TestDelete() {
	s.TestAdd()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, err := s.client.Delete(ctx, &promptv1.DeleteRequest{
		Id: 1,
	})
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Res)

	var got dao.Prompt
	err = s.db.Where("id = ?", 1).First(&got).Error
	require.NoError(s.T(), err)
	require.Equal(s.T(), uint8(0), got.Status)
	require.True(s.T(), got.Utime > got.Ctime)
}
