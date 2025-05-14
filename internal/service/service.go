package service

import (
	"context"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm"
)

type AIService struct {
	handler llm.LLMHandler
}

func NewAIService(handler llm.LLMHandler) *AIService {
	return &AIService{handler: handler}
}

func (svc *AIService) Stream(ctx context.Context, req domain.LLMRequest) (chan domain.StreamEvent, error) {
	return svc.handler.StreamHandle(ctx, req)
}

func (svc *AIService) Invoke(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error) {
	return svc.handler.Handle(ctx, req)
}
