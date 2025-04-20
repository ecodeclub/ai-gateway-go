package test

import (
	"context"
	ai "github.com/ecodeclub/ai-gateway-go/api/gen/ai/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	igrpc "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/test/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ServerTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	handler *mocks.MockLLMHandler
	server  *igrpc.Server
}

func TestServer(t *testing.T) {
	suite.Run(t, &ServerTestSuite{})
}

func (s *ServerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.handler = mocks.NewMockLLMHandler(s.ctrl)
	svc := service.NewAIService(s.handler)
	s.server = igrpc.NewServer(svc)
}

func (s *ServerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ServerTestSuite) TestStream() {
	t := s.T()

	testcases := []struct {
		name   string
		before func()
		want   []domain.StreamEvent
	}{
		{
			name: "stream event",
			before: func() {
				streamChan := make(chan domain.StreamEvent, 2)
				streamChan <- domain.StreamEvent{Content: "event1"}
				streamChan <- domain.StreamEvent{Content: "event2"}
				close(streamChan)
				s.handler.EXPECT().StreamHandle(gomock.Any(), gomock.Any()).Return(streamChan, nil)
			},
			want: []domain.StreamEvent{{Content: "event1"}, {Content: "event2"}},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			mockStream := &mocks.MockStreamServer{Ctx: context.Background()}
			err := s.server.Stream(&ai.LLMRequest{}, mockStream)
			require.NoError(t, err)

			for i, event := range tc.want {
				assert.Equal(t, event.Content, mockStream.Events[i].Content)
			}
		})
	}
}

func (s *ServerTestSuite) TestInvoke() {
	t := s.T()

	testcases := []struct {
		name   string
		before func()
		want   domain.LLMResponse
	}{
		{
			name: "stream event",
			before: func() {
				resp := domain.LLMResponse{Content: "event1"}
				s.handler.EXPECT().Handle(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			want: domain.LLMResponse{Content: "event1"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			invoke, err := s.server.Invoke(context.Background(), &ai.LLMRequest{Id: "1", Text: "hello"})
			require.NoError(t, err)
			require.Equal(t, tc.want.Content, invoke.Content)
		})
	}
}
