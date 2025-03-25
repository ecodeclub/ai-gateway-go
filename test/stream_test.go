package test

import (
	"context"
	"errors"
	ds "github.com/cohesion-org/deepseek-go"
	pb "github.com/ecodeclub/ai-gateway-go/api/proto"
	GRPC "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/platform/deepseek"
	"github.com/gotomicro/ego"
	cegrpc "github.com/gotomicro/ego/client/egrpc"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"github.com/gotomicro/ego/server/egrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"io"
	"os"
	"testing"
)

var token = os.Getenv("DEEPSEEK_TOKEN")

type StreamTest struct {
	suite.Suite
	client pb.AIServiceClient
}

func TestAIService(t *testing.T) {
	suite.Run(t, &StreamTest{})
}

func AIServer() server.Server {
	handler := deepseek.NewHandler(ds.NewClient(token))
	svc := service.NewAIService(handler)
	build := egrpc.Load("server.grpc").Build()
	pb.RegisterAIServiceServer(build.Server, GRPC.NewServer(svc))
	return build
}

func (s *StreamTest) SetupSuite() {
	go func() {
		if err := ego.New().Serve(AIServer()).Run(); err != nil {
			elog.Panic("startup", elog.Any("err", err))
		}
	}()

	grpcConn := cegrpc.Load("client.grpc").Build()
	s.client = pb.NewAIServiceClient(grpcConn.ClientConn)
}

func (s *StreamTest) TestStream() {
	t := s.T()
	tests := []struct {
		Name string
		Id   string
		Text string
	}{
		{Name: "英文问答", Id: "1", Text: "Hello"},
		{Name: "中文问答", Id: "2", Text: "你好"},
	}

	for _, tt := range tests {
		t.Run(tt.Id, func(t *testing.T) {
			stream, err := s.client.Stream(context.Background(), &pb.StreamRequest{Id: tt.Id, Text: tt.Text})
			assert.NoError(t, err)
			s.stream(t, stream)
		})
	}
}

func (s *StreamTest) stream(t *testing.T, stream grpc.ServerStreamingClient[pb.StreamResponse]) {
	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			assert.NoError(t, err)
		}
		// 表示结束
		if resp.Final {
			break
		}

		assert.NotEmpty(t, resp.Text)
		t.Log(resp.Text)
	}
}
