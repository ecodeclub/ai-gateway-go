package main

import (
	"context"
	"errors"
	ai "github.com/ecodeclub/ai-gateway-go/api/proto/gen/api/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"io"

	"github.com/stretchr/testify/suite"
	"testing"
)

type AIServiceSuite struct {
	suite.Suite
	client ai.AIServiceClient
}

func (as *AIServiceSuite) SetupSuite() {
	t := as.T()
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())

	require.NoError(t, err)
	as.client = ai.NewAIServiceClient(conn)
}

func (as *AIServiceSuite) TestStream() {
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
		stream, err := as.client.Stream(
			context.Background(),
			&ai.StreamRequest{Id: tc.Id, Text: tc.text})

		require.NoError(t, err)

		for {
			resp, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				require.NoError(t, err)
			}

			if resp.Final == true {
				return
			}

			assert.Empty(t, resp.Err)
			t.Log(resp.Content)
		}
	}
}

func TestAIService(t *testing.T) {
	suite.Run(t, &AIServiceSuite{})
}
