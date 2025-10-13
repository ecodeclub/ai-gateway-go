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
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/loadcfg"
	iopenai "github.com/ecodeclub/ai-gateway-go/internal/service/stream/openai"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/render"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/store"
	"github.com/gotomicro/ego/core/econf"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func InitOpenAIHandler(registry *fcall.Registry) *iopenai.Handler {
	type OpenAIConfig struct {
		APIKey  string `json:"apiKey"`
		BaseURL string `json:"baseURL"`
		// 一些公共的头部
		Headers map[string]string `json:"headers"`
	}
	var cfg OpenAIConfig

	err := econf.UnmarshalKey("openai", &cfg)
	if err != nil {
		panic(err)
	}
	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
	)
	return iopenai.NewHandler(client, registry, cfg.Headers)
}

func InitStreamHandler(
	loadcfgHdl *loadcfg.RebuildContextHandler,
	renderHdl *render.Handler,
	storeHdl *store.Handler,
	openaiHdl *iopenai.Handler,
) stream.Handler {
	// 组装各个 next
	// 当前顺序 loadcfg -> render -> store -> openai
	// openai 里面还要发起下一次调用，于是重归 loadcfg
	loadcfgHdl.Next = renderHdl
	renderHdl.Next = storeHdl
	storeHdl.Next = openaiHdl
	openaiHdl.Handler = loadcfgHdl
	return loadcfgHdl
}

func InitFuncCall(
	f3 *kbase.RAG,
	f4 *analyzer.AnalysisDialogFCall,
	f5 *savedoc.FCall,
) []fcall.FunctionCall {
	return []fcall.FunctionCall{f3, f4, f5}
}

func InitKBaseRAG() *kbase.RAG {
	type KBaseRAGConfig struct {
		URL string `json:"url" yaml:"url"`
	}
	var cfg KBaseRAGConfig
	err := econf.UnmarshalKey("kbase", &cfg)
	if err != nil {
		panic(err)
	}
	return kbase.NewKBaseRAG(cfg.URL)
}
