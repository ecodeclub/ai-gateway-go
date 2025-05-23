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
}

func TestServer(t *testing.T) {
	suite.Run(t, &ServerTestSuite{})
}

func (s *ServerTestSuite) TestStream() {
	t := s.T()

	testcases := []struct {
		name   string
		before func(handler *mocks.MockLLMHandler)
		want   []domain.StreamEvent
	}{
		{
			name: "stream event",
			before: func(handler *mocks.MockLLMHandler) {
				streamChan := make(chan domain.StreamEvent, 2)
				streamChan <- domain.StreamEvent{Content: "event1"}
				streamChan <- domain.StreamEvent{Content: "event2"}
				close(streamChan)
				handler.EXPECT().StreamHandle(gomock.Any(), gomock.Any()).Return(streamChan, nil)
			},
			want: []domain.StreamEvent{{Content: "event1"}, {Content: "event2"}},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mocks.NewMockLLMHandler(ctrl)
			svc := service.NewAIService(handler)
			server := igrpc.NewServer(svc)

			tc.before(handler)
			mockStream := &mocks.MockStreamServer{Ctx: context.Background()}
			err := server.Stream(&ai.LLMRequest{}, mockStream)
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
		before func(handler *mocks.MockLLMHandler)
		want   domain.LLMResponse
	}{
		{
			name: "stream event",
			before: func(handler *mocks.MockLLMHandler) {
				resp := domain.LLMResponse{Content: "event1"}
				handler.EXPECT().Handle(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			want: domain.LLMResponse{Content: "event1"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			handler := mocks.NewMockLLMHandler(ctrl)
			svc := service.NewAIService(handler)
			server := igrpc.NewServer(svc)
			tc.before(handler)
			invoke, err := server.Invoke(context.Background(), &ai.LLMRequest{Id: "1", Text: "hello"})
			require.NoError(t, err)
			require.Equal(t, tc.want.Content, invoke.Content)
		})
	}
}
