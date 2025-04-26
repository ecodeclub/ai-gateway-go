package mocks

import (
	"context"
	ai "github.com/ecodeclub/ai-gateway-go/api/gen/ai/v1"
)

type MockStreamServer struct {
	ai.AIService_StreamServer
	Ctx    context.Context
	Events []*ai.StreamEvent
	err    error
}

func (m *MockStreamServer) Context() context.Context {
	return m.Ctx
}

func (m *MockStreamServer) Send(event *ai.StreamEvent) error {
	m.Events = append(m.Events, event)
	return m.err
}
