package llm

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
)

type LLMHandler interface {
	StreamHandle(ctx context.Context, req domain.LLMRequest) (chan domain.StreamEvent, error)
	Handle(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error)
}
