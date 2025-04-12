package main

import (
	"context"
	"errors"
	ds "github.com/cohesion-org/deepseek-go"
	ai "github.com/ecodeclub/ai-gateway-go/api/gen/ai/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/platform/deepseek"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/client/egrpc"
	"github.com/gotomicro/ego/core/econf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
	"time"
)

type AIServiceSuite struct {
	suite.Suite
	service *service.AIService
	client  ai.AIServiceClient
}

func (as *AIServiceSuite) SetupSuite() {
	token := econf.GetString("deepseek.token")
	handler := deepseek.NewHandler(ds.NewClient(token))
	as.service = service.NewAIService(handler)

	econf.Set("grpc.client.addr", "127.0.0.1:9002")
	grpcConn := egrpc.Load("grpc.client").Build()
	as.client = ai.NewAIServiceClient(grpcConn.ClientConn)
	err := ego.New().Invoker(as.TestStream).Run()
	require.NoError(as.T(), err)
}

func (as *AIServiceSuite) TestStream() error {
	t := as.T()

	testCases := []struct {
		name string
		Id   string
		text string
	}{
		{
			name: "hello",
			Id:   "1",
			text: "hello, deepseek",
		},
		{
			name: "你好",
			Id:   "2",
			text: "你好, deepseek",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stream, err := as.client.Stream(
				context.Background(),
				&ai.LLMRequest{Id: tc.Id, Text: tc.text})

			require.NoError(t, err)

			var answer = ""
			for {
				resp, err := stream.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					require.NoError(t, err)
				}

				if resp.Final == true {
					break
				}

				assert.Empty(t, resp.Err)
				answer += resp.Content
			}
			assert.Contains(t, answer, "DeepSeek")
		})
	}
	return nil
}

func (as *AIServiceSuite) TestInvoke() error {
	t := as.T()

	testCases := []struct {
		name string
		Id   string
		text string
	}{
		{
			name: "hello",
			Id:   "1",
			text: "hello, deepseek",
		},
		{
			name: "你好",
			Id:   "2",
			text: "你好, deepseek",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			resp, err := as.service.Invoke(
				ctx,
				domain.LLMRequest{Id: "1", Text: "hello"})

			require.NoError(t, err)
			assert.Contains(t, resp.Content, "Hello")
		})
	}

	return nil
}

func TestAIService(t *testing.T) {
	suite.Run(t, &AIServiceSuite{})
}
