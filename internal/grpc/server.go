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

package grpc

import (
	"context"

	ai "github.com/ecodeclub/ai-gateway-go/api/gen/ai/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
)

type Server struct {
	svc *service.AIService
	ai.UnimplementedAIServiceServer
}

func NewServer(svc *service.AIService) *Server {
	return &Server{svc: svc}
}

func (server *Server) Invoke(ctx context.Context, r *ai.LLMRequest) (*ai.LLMResponse, error) {
	resp, err := server.svc.Invoke(
		ctx,
		domain.LLMRequest{Id: r.GetId(), Text: r.GetText(), ContentType: domain.ContentType(r.GetContentType())})

	if err != nil {
		return &ai.LLMResponse{}, err
	}

	return &ai.LLMResponse{Content: resp.Content}, nil
}

func (server *Server) Stream(r *ai.LLMRequest, resp ai.AIService_StreamServer) error {
	ctx := resp.Context()

	ch, err := server.svc.Stream(
		ctx,
		domain.LLMRequest{Id: r.GetId(), Text: r.GetText(), ContentType: domain.ContentType(r.GetContentType())})

	if err != nil {
		return err
	}

	return server.stream(ctx, ch, resp)
}

func (server *Server) stream(ctx context.Context, ch chan domain.StreamEvent, resp ai.AIService_StreamServer) error {
	var err error
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e, ok := <-ch:
			if !ok || e.Done {
				err = resp.Send(&ai.StreamEvent{Final: true})
				return err
			}
			if e.Error != nil {
				err = resp.Send(&ai.StreamEvent{Err: e.Error.Error()})
				return err
			}
			err = resp.Send(&ai.StreamEvent{Final: false, Content: e.Content})
			if err != nil {
				return err
			}
		}
	}
}
