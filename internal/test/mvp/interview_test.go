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

package mvp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	chatv1 "github.com/ecodeclub/ai-gateway-go/api/proto/gen/chat/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	igrpc "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall/savedoc"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/loadcfg"
	openaistream "github.com/ecodeclub/ai-gateway-go/internal/service/stream/openai"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/render"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/store"
	_ "github.com/ecodeclub/ai-gateway-go/internal/test"
	testioc "github.com/ecodeclub/ai-gateway-go/internal/test/ioc"
	openai3 "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestGrpcServer 启动 gRPC 服务器用于面试功能测试
// 端口: 9090
// 功能: 提供 StreamV1 接口，处理面试逻辑
func TestGrpcServer(t *testing.T) {
	ctx := context.Background()

	// 1. 初始化 stream handler（需要 OpenAI client）
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Fatal("未设置 API_KEY 环境变量")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		t.Fatal("未设置 BASE_URL 环境变量")
	}

	internalToken := os.Getenv("INTERNAL_TOKEN")
	if internalToken == "" {
		t.Fatal("未设置 INTERNAL_TOKEN 环境变量")
	}

	origin := os.Getenv("ORIGIN")
	if origin == "" {
		t.Fatal("未设置 ORIGIN 环境变量")
	}

	// 获取额外的 headers
	headers := make(map[string]string)
	headers["X-Internal-Token"] = internalToken
	headers["Origin"] = origin

	// 创建 OpenAI client
	client := openai3.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	// 2. 初始化测试应用（使用 IOC）
	registry := fcall.NewFunctionCallRegistry()

	// 注册 save_doc function call
	saveDocFCall := savedoc.NewFCall()
	registry.Register(saveDocFCall)

	openaiHandler := openaistream.NewHandler(client, registry, headers)

	// 需要先创建依赖的 repositories
	db := testioc.InitDB()
	chatRepo := repository.NewChatRepo(dao.NewChatDAO(db), nil)
	providerDAO := dao.NewProviderDAO(db)
	providerRepo := repository.NewProviderRepository(providerDAO)
	invConfigDAO := dao.NewInvocationConfigDAO(db)
	invConfigRepo := repository.NewInvocationConfigRepo(invConfigDAO, providerRepo)

	// 创建 stream handler
	streamHandler := initStreamHandler(openaiHandler, chatRepo, invConfigRepo, providerRepo)

	app := testioc.InitApp(testioc.TestOnly{
		Handler: streamHandler,
	})

	// 3. 准备测试数据
	log.Println("📝 准备测试数据...")

	// 3.1 创建 Provider（OpenAI）
	providerID, err := providerDAO.SaveProvider(ctx, dao.Provider{
		Name:   "OpenAI",
		APIKey: "test-api-key", // 测试环境使用占位符
	})
	if err != nil {
		t.Fatalf("创建 Provider 失败: %v", err)
	}
	log.Printf("   ✓ 创建 Provider: OpenAI (ID: %d)", providerID)

	// 3.2 创建 Model（gpt-4o-mini）
	modelID, err := providerDAO.SaveModel(ctx, dao.Model{
		Name:        "gpt-4o-mini",
		Pid:         providerID,
		InputPrice:  150, // $0.150 / 1M tokens
		OutputPrice: 600, // $0.600 / 1M tokens
		PriceMode:   "token",
	})
	if err != nil {
		t.Fatalf("创建 Model 失败: %v", err)
	}
	log.Printf("   ✓ 创建 Model: gpt-4o-mini (ID: %d)", modelID)

	// 3.3 创建 BizConfig
	bizConfigDAO := dao.NewBizConfigDAO(db)
	bizRepo := repository.NewBizConfigRepository(bizConfigDAO)
	bizSvc := service.NewBizConfigService(bizRepo)

	bizID, err := bizSvc.Save(ctx, domain.BizConfig{
		Name:      "面试测试",
		OwnerID:   1,
		OwnerType: "user",
	})
	if err != nil {
		t.Fatalf("创建 BizConfig 失败: %v", err)
	}
	log.Printf("   ✓ 创建 BizConfig: 面试测试 (ID: %d)", bizID)

	// 3.4 创建 InvocationConfig
	invSvc := service.NewInvocationConfigService(invConfigRepo, bizRepo, providerRepo)

	cfgID, err := invSvc.Save(ctx, domain.InvocationConfig{
		ID:          100001,
		Name:        "MySQL模拟面试助手",
		Biz:         domain.BizConfig{ID: bizID},
		Description: "用于MySQL模拟面试的配置",
	})
	if err != nil {
		t.Fatalf("创建 InvocationConfig 失败: %v", err)
	}
	log.Printf("   ✓ 创建 InvocationConfig: MySQL模拟面试助手 (ID: %d)", cfgID)

	// 3.5 创建 InvocationConfigVersion（active）
	versionID, err := invSvc.SaveVersion(ctx, domain.InvocationConfigVersion{
		Config:       domain.InvocationConfig{ID: cfgID},
		Model:        domain.Model{ID: modelID},
		Version:      "v1.0",
		Status:       domain.InvocationCfgVersionStatusActive,
		SystemPrompt: "你是专业的MySQL面试官助手，负责出题、评分和总结。",
		Prompt: `你是MySQL面试官，负责出题、评分和总结。

用户输入：{{.Input}}

# 当前面试历史
{{if .InterviewHistory}}
{{$history := fromJson .InterviewHistory}}
已答题目：
{{range $record := $history}}
- {{$record.question}}
  回答：{{$record.answer}}
  评分：内容{{$record.scores.content_score}} 完整性{{$record.scores.coverage_score}} 结构{{$record.scores.structure_score}}
{{end}}
{{else}}
（无历史记录）
{{end}}

# 你的任务
1. 如果用户说"开始面试"，**只返回题目JSON，不要调用任何 function**
2. 如果用户提供了答案，**先返回评分JSON（此时包含下一题）**给用户，然后再调用 save_doc 函数将评分（此时不包含下一题内容）返回给开发者保存用户历史。
3. 如果用户说"结束面试"，**只返回总结JSON，不要调用任何 function**

# 输出格式（严格遵守）

## 场景1 - 出题（不要调用 function）：

{
  "type": "question",
  "question": "什么是MySQL索引？请简述其作用。"
}
注意：出题时只返回 JSON，不要调用 save_doc 函数

## 场景2 - 评分（评分后必须调用 save_doc）：

1. 先返回JSON给用户，包含下一题内容：

{
  "type": "evaluation",
  "scores": {
    "content_score": 85,
    "coverage_score": 78,
    "structure_score": 90
  },
  "evaluation": {
    "key_points_hit": ["B+树", "查询加速"],
    "missed_points": ["索引失效场景"],
    "suggestion": "可以补充索引失效的场景"
  },
  "next_question": "什么是事务？请列举ACID特性。"
}

2. 再调用 save_doc 函数返回评分给开发者保存历史：

- varName: "InterviewHistory"
- type: "json"
- content: 完整的历史数组JSON字符串，包含所有已评分的题目
    - 例如：如果这是第一题，content 应该是：[{"question":"什么是MySQL索引？","answer":"用户的回答内容","scores":{"content_score":85,"coverage_score":78,"structure_score":90},"evaluation":{"key_points_hit":["B+树","查询加速"],"missed_points":["索引失效场景"],"suggestion":"可以补充索引失效的场景"}}]
    - 如果已有历史，则追加新记录到数组末尾。

## 场景3 - 总结（不要调用 function）：

{
  "type": "summary",
  "overall_score": 82,
  "strengths": ["基础扎实", "表达清晰"],
  "weaknesses": ["高级特性欠缺"],
  "priority_actions": ["深入学习锁机制", "实践索引优化"]
}

注意：总结时只返回 JSON，不要调用 save_doc

# 重要规则：
1. 只返回JSON，不要任何其他文字
2. 在场景2（评分）要返回两个内容，一个是带下一题内容的，另一个是通过 save_doc 函数返回的评价。一定是返回两次
3. 场景1（出题）和场景3（总结）都不要调用任何 function`,
		Temperature: 0.7,
		TopP:        1.0,
		MaxTokens:   2000,
		Functions: []domain.Function{
			{
				Name: "save_doc",
				Definition: `{
  "name": "save_doc",
  "description": "保存完整的面试历史记录到变量中，每次评估完回答后必须调用",
  "parameters": {
    "type": "object",
    "properties": {
      "varName": {
        "type": "string",
        "enum": ["InterviewHistory"],
        "description": "变量名，固定为 InterviewHistory"
      },
      "type": {
        "type": "string",
        "enum": ["json"],
        "description": "数据类型，固定为 json"
      },
      "content": {
        "type": "string",
        "description": "完整的面试历史，JSON数组字符串。每个元素包含: question(题目), answer(回答), scores(评分对象), evaluation(评价对象)。必须包含之前的所有记录加上当前新记录。"
      }
    },
    "required": ["varName", "content", "type"]
  }
}`,
			},
		},
	})
	if err != nil {
		t.Fatalf("创建 InvocationConfigVersion 失败: %v", err)
	}
	log.Printf("   ✓ 创建 InvocationConfigVersion: v1.0 (ID: %d, Status: active)", versionID)

	// 清理函数
	defer func() {
		log.Println("\n🧹 清理测试数据...")
		db.Delete(&dao.InvocationConfigVersion{}, versionID)
		db.Delete(&dao.InvocationConfig{}, cfgID)
		db.Delete(&dao.BizConfig{}, bizID)
		db.Delete(&dao.Model{}, modelID)
		db.Delete(&dao.Provider{}, providerID)
		log.Println("   ✓ 测试数据已清理")
	}()

	log.Println("\n✅ 数据准备完成，测试环境已就绪")

	// 4. 启动 gRPC 服务器
	chatSvc := app.ChatService
	chatServer := igrpc.NewChatServer(chatSvc, streamHandler)

	grpcServer := grpc.NewServer()
	chatv1.RegisterServiceServer(grpcServer, chatServer)

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		t.Fatalf("监听端口失败: %v", err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC 服务器错误: %v", err)
		}
	}()

	defer grpcServer.Stop()

	log.Println("🚀 gRPC 服务器启动于 :9090")
	log.Printf("📋 InvocationConfig ID: %d", cfgID)
	log.Println("💡 按 Ctrl+C 停止服务器")
	log.Println("---")

	// 保持运行
	select {}
}

// initStreamHandler 初始化 stream handler 链
func initStreamHandler(
	openaiHdl *openaistream.Handler,
	chatRepo *repository.ChatRepo,
	invConfigRepo *repository.InvocationConfigRepo,
	providerRepo *repository.ProviderRepository,
) stream.Handler {
	loadcfgHdl := loadcfg.NewLoadConfigHandler(chatRepo, invConfigRepo, providerRepo)
	renderHdl := render.NewHandler()
	storeHdl := store.NewHandler(chatRepo)

	// 组装责任链: loadcfg -> render -> store -> openai
	loadcfgHdl.Next = renderHdl
	renderHdl.Next = storeHdl
	storeHdl.Next = openaiHdl
	openaiHdl.Handler = loadcfgHdl

	return loadcfgHdl
}

// TestInterviewProxyServer 启动 HTTP → gRPC 代理服务器
// 端口: 8080
// 功能: 将前端 HTTP 请求转换为 gRPC 调用，并将 gRPC 流式响应转换为 SSE
func TestInterviewProxyServer(t *testing.T) {
	// 1. 连接到 gRPC 服务器
	conn, err := grpc.Dial("localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("连接 gRPC 失败: %v", err)
	}
	defer conn.Close()

	client := chatv1.NewServiceClient(conn)

	mux := http.NewServeMux()

	// 2. CORS 处理函数
	corsHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next(w, r)
		}
	}

	// 3. 创建 Chat 接口
	mux.HandleFunc("/api/interview/chat/create", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Uid   int64  `json:"uid"`
			Title string `json:"title"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("📝 创建 Chat: uid=%d, title=%s", req.Uid, req.Title)

		// 调用 gRPC Save
		resp, err := client.Save(context.Background(), &chatv1.SaveRequest{
			Chat: &chatv1.Chat{
				Uid:   req.Uid,
				Title: req.Title,
			},
		})
		if err != nil {
			log.Printf("❌ 创建 Chat 失败: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("✅ Chat 已创建: %s", resp.Sn)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"chat_sn": resp.Sn,
		})
	}))

	// 4. StreamV1 流式接口（SSE）
	mux.HandleFunc("/api/interview/stream", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ChatSn   string `json:"chat_sn"`
			Input    string `json:"input"`
			ConfigId int64  `json:"config_id"`
			Uid      int64  `json:"uid"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("📥 收到请求: chat_sn=%s, input=%s (前30字)", req.ChatSn, truncate(req.Input, 30))

		// 调用 gRPC StreamV1
		stream, err := client.StreamV1(context.Background(), &chatv1.StreamV1Request{
			ChatSn: req.ChatSn,
			Input: &chatv1.UserInput{
				Content: req.Input,
			},
			InvocationConfigId: req.ConfigId,
			Uid:                req.Uid,
			Key:                "", // 未使用，传空字符串
		})
		if err != nil {
			log.Printf("❌ 调用 StreamV1 失败: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 设置 SSE 响应头
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// 转发流式响应
		deltaCount := 0
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				fmt.Fprintf(w, "event: done\ndata: {}\n\n")
				flusher.Flush()
				log.Printf("✅ Stream 完成 (共 %d 个 Delta 事件)", deltaCount)
				break
			}
			if err != nil {
				log.Printf("❌ Stream 错误: %v", err)
				fmt.Fprintf(w, "event: error\ndata: {\"message\": \"%s\"}\n\n", err.Error())
				flusher.Flush()
				break
			}

			// 转换为 JSON 并发送
			data, _ := json.Marshal(resp)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

			// 日志（Delta 事件）
			if resp.GetDelta() != nil {
				deltaCount++
				log.Printf("📤 Delta #%d: %s", deltaCount, resp.GetDelta().Content)
			}
		}
	}))

	// 5. 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"grpc":   "localhost:9090",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	log.Println("🚀 HTTP 代理服务器启动于 :8080")
	log.Println("📡 转发目标: localhost:9090 (gRPC)")
	log.Println("📍 端点:")
	log.Println("   - POST /api/interview/chat/create")
	log.Println("   - POST /api/interview/stream")
	log.Println("   - GET  /health")
	log.Println("💡 按 Ctrl+C 停止服务器")
	log.Println("---")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		t.Fatalf("HTTP 服务器启动失败: %v", err)
	}
}

// TestAudioProxyServer 启动音频转文本代理服务器
// 端口: 8000
// 功能: 代理音频转文本请求到 OpenAI API
func TestAudioProxyServer(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Fatal("未设置 API_KEY 环境变量")
	}

	internalToken := os.Getenv("INTERNAL_TOKEN")
	if internalToken == "" {
		t.Fatal("未设置 INTERNAL_TOKEN 环境变量")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		t.Fatal("未设置 INTERNAL_TOKEN 环境变量")
	}

	origin := os.Getenv("ORIGIN")
	if origin == "" {
		t.Fatal("未设置 ORIGIN 环境变量")
	}

	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "8000"
	}

	mux := http.NewServeMux()

	// ============ 通用代理处理器 ============
	proxyHandler := func(w http.ResponseWriter, r *http.Request) {
		// 启用 CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 读取请求体
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取请求体失败: %v", err), http.StatusBadRequest)
			return
		}
		r.Body.Close()

		// 构建目标 URL
		path := strings.TrimPrefix(r.URL.Path, "/api")
		targetURL := baseURL + path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		log.Printf("🔄 代理请求: %s %s -> %s", r.Method, r.URL.Path, targetURL)
		if len(bodyBytes) > 0 && len(bodyBytes) < 2000 {
			log.Printf("📤 请求体: %s", string(bodyBytes))
		}

		// 创建新请求
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, r.Method, targetURL, bytes.NewReader(bodyBytes))
		if err != nil {
			http.Error(w, fmt.Sprintf("创建请求失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 复制请求头
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-Internal-Token", internalToken)
		req.Header.Set("Origin", origin)
		req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		// 发送请求
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("请求失败: %v", err), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 读取响应
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取响应失败: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("✅ 响应状态: %d, 大小: %d 字节", resp.StatusCode, len(respBytes))
		if len(respBytes) < 2000 {
			log.Printf("📥 响应体: %s", string(respBytes))
		}

		// 复制响应头（跳过 CORS 头，避免重复）
		skipHeaders := map[string]bool{
			"Access-Control-Allow-Origin":      true,
			"Access-Control-Allow-Methods":     true,
			"Access-Control-Allow-Headers":     true,
			"Access-Control-Allow-Credentials": true,
			"Access-Control-Expose-Headers":    true,
			"Access-Control-Max-Age":           true,
		}

		for k, v := range resp.Header {
			if skipHeaders[k] {
				continue
			}
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}

		w.WriteHeader(resp.StatusCode)
		w.Write(respBytes)
	}

	// ============ 注册路由 ============
	// Audio API
	mux.HandleFunc("/api/audio/transcriptions", proxyHandler)

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"base_url": baseURL,
			"time":     time.Now().Format(time.RFC3339),
		})
	})

	addr := ":" + port
	log.Printf("🚀 远程代理服务器启动于 http://localhost%s", addr)
	log.Printf("📡 转发目标: %s", baseURL)
	log.Printf("✅ 支持的端点:")
	log.Printf("   - POST /api/audio/transcriptions")
	log.Printf("   - GET  /health")
	log.Printf("💡 提示: 按 Ctrl+C 停止服务器")
	log.Println("---")

	if err := http.ListenAndServe(addr, mux); err != nil {
		t.Fatalf("服务器启动失败: %v", err)
	}
}

// truncate 截断字符串用于日志显示
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
