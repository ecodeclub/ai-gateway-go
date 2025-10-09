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

package ioc

import (
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall/analyzer"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall/kbase"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall/savedoc"
	iopenai "github.com/ecodeclub/ai-gateway-go/internal/service/stream/openai"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/rebuildctx"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/render"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/store"
	"github.com/gotomicro/ego/core/econf"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

func InitOpenAIClient() openai.Client {
	type OpenAIConfig struct {
		APIKey  string `json:"apiKey"`
		BaseURL string `json:"baseURL"`
	}
	var cfg OpenAIConfig

	err := econf.UnmarshalKey("openai", &cfg)
	if err != nil {
		panic(err)
	}
	return openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
	)
}

func InitStreamHandler(
	rebuildHdl *rebuildctx.RebuildContextHandler,
	renderHdl *render.Handler,
	storeHdl *store.Handler,
	openaiHdl *iopenai.Handler,
) stream.Handler {
	// 组装各个 next
	// 当前顺序 rebuild -> render -> store -> openai
	rebuildHdl.Next = renderHdl
	renderHdl.Next = storeHdl
	storeHdl.Next = openaiHdl
	return rebuildHdl
}

func InitFuncCall(
	f3 *kbase.RAG,
	f4 *analyzer.AnalysisDialogFCall,
	f5 *savedoc.FCall,
) []fcall.FunctionCall {
	return []fcall.FunctionCall{f3, f4, f5}
}

func InitKBaseRAG(base *fcall.BaseFCall) *kbase.RAG {
	type KBaseRAGConfig struct {
		URL string `json:"url" yaml:"url"`
	}
	var cfg KBaseRAGConfig
	err := econf.UnmarshalKey("kbase", &cfg)
	if err != nil {
		panic(err)
	}
	return kbase.NewKBaseRAG(cfg.URL, base)
}
