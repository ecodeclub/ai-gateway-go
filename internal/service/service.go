// Copyright 2021 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm"
)

type AIService struct {
	handler llm.Handler
}

func NewAIService(handler llm.Handler) *AIService {
	return &AIService{handler: handler}
}

func (svc *AIService) Stream(ctx context.Context, req domain.Message) (chan domain.StreamEvent, error) {
	return svc.handler.StreamHandle(ctx, []domain.Message{req})
}

func (svc *AIService) Invoke(ctx context.Context, req domain.Message) (domain.ChatResponse, error) {
	return svc.handler.Handle(ctx, []domain.Message{req})
}
