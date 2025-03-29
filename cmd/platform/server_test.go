package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ecodeclub/ai-gateway-go/api/ai"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"io"

	"github.com/stretchr/testify/suite"
	"testing"
)

type AIServiceSuite struct {
	suite.Suite
}

func (as *AIServiceSuite) CallStream() error {
	t := as.T()

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		fmt.Print(err)
	}
	defer conn.Close()

	client := ai.NewAIServiceClient(conn)
	stream, err := client.Stream(context.Background(), &ai.StreamRequest{Id: "1", Text: "你好"})
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if resp.Final == true {
			return nil
		}

		if resp.Err != "" {
			return fmt.Errorf(resp.Err)

		}
		t.Log(resp.Text)
	}
}

func (as *AIServiceSuite) TestServer() {
	err := as.CallStream()
	require.NoError(as.T(), err)
}

func TestAIService(t *testing.T) {
	suite.Run(t, &AIServiceSuite{})
}
