// Copyright 2025 ecodeclub
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

package orchestrator

import (
	"context"
	"fmt"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/gotomicro/ego/core/elog"
)

type Orchestrator struct {
	Handler stream.Handler
	logger  *elog.Component
}

func NewOrchestrator(handler stream.Handler) *Orchestrator {
	return &Orchestrator{Handler: handler, logger: elog.DefaultLogger.With(elog.FieldComponent("Orchestrator"))}
}

func (o *Orchestrator) Stream(ctx context.Context,
	chat domain.Chat,
	sender domain.StreamEventSender) error {
	// 从 main 开始调度
	thread := chat.BizOrchestration.Main
	state := chat.CurrentTurn().StartState
	if state != "" {
		var ok bool
		thread, ok = chat.BizOrchestration.Threads[state]
		if !ok {
			return fmt.Errorf("找不到 Thread，初始 state %s", state)
		}
	}
	for {
		chat.CurrentTurn().AssistantRun.StartLLMStep(thread)

		resp, err1 := o.Handler.Stream(&domain.StreamContext{
			Ctx:    ctx,
			Sender: sender,
			Chat:   chat,
		})
		if err1 != nil {
			return err1
		}

		state := resp.NextState
		o.logger.Debug("", elog.String("NextState", state))
		if state == "" {
			return nil
		}
		var ok bool
		thread, ok = chat.BizOrchestration.Threads[state]
		if !ok {
			return fmt.Errorf("找不到 Thread %s", state)
		}
	}
}
