package main

import (
	"context"
	"errors"
	ai "github.com/ecodeclub/ai-gateway-go/api/gen/ai/v1"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/client/egrpc"
	"github.com/gotomicro/ego/core/econf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
)

type AIServiceSuite struct {
	suite.Suite
	client ai.AIServiceClient
}

func (as *AIServiceSuite) SetupSuite() {
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

	var answer = ""

	for _, tc := range testCases {
		stream, err := as.client.Stream(
			context.Background(),
			&ai.StreamRequest{Id: tc.Id, Text: tc.text})

		require.NoError(t, err)

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
	}
	return nil
}

func TestAIService(t *testing.T) {
	suite.Run(t, &AIServiceSuite{})
}
