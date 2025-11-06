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

package multifunc

import (
	"encoding/json"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/openai/openai-go/v3/responses"
)

type FCall struct {
	Registry *fcall.Registry
}

func NewFCall() *FCall {
	return &FCall{}
}

func (c *FCall) Name() string {
	return "multi_call"
}

func (c *FCall) Call(ctx *domain.StreamContext, req fcall.Request) (fcall.Response, error) {
	var fcReq Request
	err := json.Unmarshal(req.Args, &fcReq)
	if err != nil {
		return fcall.Response{}, err
	}
	var resp fcall.Response
	for _, call := range fcReq.Calls {
		var fc fcall.FunctionCall
		fc, err = c.Registry.Lookup(call.Name)
		if err != nil {
			return fcall.Response{}, err
		}
		resp, err = fc.Call(ctx, fcall.Request{
			Args: []byte(call.Arguments),
		})
		if err != nil {
			return fcall.Response{}, err
		}
	}
	return fcall.Response{
		Content:   fcReq.Content,
		NextState: resp.NextState,
	}, nil
}

type Request struct {
	Content   string                               `json:"content"`
	Calls     []responses.ResponseFunctionToolCall `json:"calls"`
	NextState string                               `json:"nextState"`
}
