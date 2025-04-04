package llm

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
)

type LLMHandler interface {
	StreamHandle(ctx context.Context, req domain.StreamRequest) (chan domain.StreamEvent, error)
}
