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

func (svc *AIService) Stream(ctx context.Context, s domain.StreamRequest) (chan domain.StreamEvent, error) {
	return svc.handler.StreamHandle(ctx, s)
}
