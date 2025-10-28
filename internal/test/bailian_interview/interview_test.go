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

package bailian_interview

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
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
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/bailian"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall/forward"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall/kbase"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall/savedoc"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/loadcfg"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/render"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/store"
	_ "github.com/ecodeclub/ai-gateway-go/internal/test"
	testioc "github.com/ecodeclub/ai-gateway-go/internal/test/ioc"
	elasticsearch "github.com/elastic/go-elasticsearch/v9"
	openai3 "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tencentyun/cos-go-sdk-v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:embed system_prompt.md
var systemPrompt string

//go:embed user_prompt.md
var userPrompt string

// TestGrpcServer 启动 gRPC 服务器用于面试功能测试
// 端口: 9090
// 功能: 提供 StreamV1 接口，处理面试逻辑
func TestGrpcServer(t *testing.T) {
	ctx := context.Background()

	// 1. 初始化 stream handler（百炼 client）
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Fatal("未设置 API_KEY 环境变量")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		t.Fatal("未设置 BASE_URL 环境变量")
	}

	// 创建百炼 client（兼容 OpenAI SDK）
	client := openai3.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	// 空的 headers（百炼不需要额外认证头）
	headers := make(map[string]string)

	// 2. 准备Kbase测试数据（ES中的面试题）
	prepareKbaseTestData(t)

	// 3. 初始化测试应用（使用 IOC）
	registry := fcall.NewFunctionCallRegistry()

	// 注册 kbase_rag function call
	kbaseRAG := kbase.NewKBaseRAG("http://localhost:8082/api/v1/es_search")
	registry.Register(kbaseRAG)

	// 注册 forward_result function call
	forwardResult := forward.NewResult()
	registry.Register(forwardResult)

	// 注册 save_doc function call
	saveDocFCall := savedoc.NewFCall()
	registry.Register(saveDocFCall)

	bailianHandler := bailian.NewHandler(client, registry, headers)

	// 需要先创建依赖的 repositories
	db := testioc.InitDB()
	chatRepo := repository.NewChatRepo(dao.NewChatDAO(db), nil)
	providerDAO := dao.NewProviderDAO(db)
	providerRepo := repository.NewProviderRepository(providerDAO)
	invConfigDAO := dao.NewInvocationConfigDAO(db)
	invConfigRepo := repository.NewInvocationConfigRepo(invConfigDAO, providerRepo)

	// 创建 stream handler
	streamHandler := initStreamHandler(bailianHandler, chatRepo, invConfigRepo, providerRepo)

	app := testioc.InitApp(testioc.TestOnly{
		Handler: streamHandler,
	})

	// 4. 准备测试数据
	log.Println("📝 准备测试数据...")

	// 4.1 创建 Provider（Bailian）
	providerID, err := providerDAO.SaveProvider(ctx, dao.Provider{
		Name:   "Bailian",
		APIKey: "test-api-key", // 测试环境使用占位符
	})
	if err != nil {
		t.Fatalf("创建 Provider 失败: %v", err)
	}
	log.Printf("   ✓ 创建 Provider: Bailian (ID: %d)", providerID)

	// 4.2 创建 Model
	modelID, err := providerDAO.SaveModel(ctx, dao.Model{
		Name:        "qwen3-max",
		Pid:         providerID,
		InputPrice:  4,  // ¥0.004 / 1K tokens
		OutputPrice: 12, // ¥0.012 / 1K tokens
		PriceMode:   "token",
	})
	if err != nil {
		t.Fatalf("创建 Model 失败: %v", err)
	}
	log.Printf("   ✓ 创建 Model: qwen3-max (ID: %d)", modelID)

	// 4.3 创建 BizConfig
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

	// 4.4 创建 InvocationConfig
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

	// 4.5 创建 InvocationConfigVersion（active）
	versionID, err := invSvc.SaveVersion(ctx, domain.InvocationConfigVersion{
		Config:       domain.InvocationConfig{ID: cfgID},
		Model:        domain.Model{ID: modelID},
		Version:      "v1.0",
		Status:       domain.InvocationCfgVersionStatusActive,
		SystemPrompt: systemPrompt, // 静态内容：状态机定义、规则、示例（可被LLM缓存）
		Prompt:       userPrompt,   // 动态内容：用户输入、历史记录（每次都不同）
		Temperature:  0,
		TopP:         1.0,
		MaxTokens:    200000,
		Functions: []domain.Function{
			{
				Name: "kbase_rag",
				Definition: `{
  "name": "kbase_rag",
  "description": "从知识库中检索资源（面试题、面经、案例等）。传入完整的Elasticsearch DSL查询对象，会直接透传给ES执行。",
  "strict": true,
  "parameters": {
    "type": "object",
    "additionalProperties": false,
    "properties": {
      "varName": {
        "type": "string",
        "description": "变量名，用于保存本次查询结果。使用 Question_N 格式命名，例如 Question_1"
      },
      "es_dsl": {
        "type": "object",
        "description": "Elasticsearch DSL 查询对象，包含 index 和 query 两个字段。",
        "additionalProperties": false,
        "properties": {
          "index": {
            "type": "string",
            "description": "ES索引名称，固定为 interview_questions_mysql",
            "enum": ["interview_questions_mysql"]
          },
          "query": {
            "type": "object",
            "description": "ES查询体，包含四个平级字段：query、size、sort、aggs",
            "additionalProperties": false,
            "properties": {
              "query": {
                "type": "object",
                "description": "ES查询体，必须包含bool查询结构",
                "additionalProperties": false,
                "properties": {
                  "bool": {
                    "type": "object",
                    "description": "布尔查询，必须包含must和must_not",
                    "additionalProperties": false,
                    "properties": {
                      "must": {
                        "type": "array",
                        "description": "必须匹配的条件",
                        "items": {
                          "type": "object",
                          "additionalProperties": false,
                          "properties": {
                            "term": {
                              "type": "object",
                              "additionalProperties": false,
                              "properties": {
                                "level": {
                                  "type": "object",
                                  "additionalProperties": false,
                                  "properties": {
                                    "value": {
                                      "type": "string",
                                      "const": "junior"
                                    }
                                  },
                                  "required": ["value"]
                                }
                              },
                              "required": ["level"]
                            }
                          },
                          "required": ["term"]
                        }
                      },
                      "must_not": {
                        "type": "array",
                        "description": "必须不匹配的条件",
                        "items": {
                          "type": "object",
                          "additionalProperties": false,
                          "properties": {
                            "terms": {
                              "type": "object",
                              "additionalProperties": false,
                              "properties": {
                                "question_id": {
                                  "type": "array",
                                  "items": {"type": "integer"},
                                  "description": "要排除的题目ID列表"
                                }
                              },
                              "required": ["question_id"]
                            }
                          },
                          "required": ["terms"]
                        }
                      }
                    },
                    "required": ["must", "must_not"]
                  }
                },
                "required": ["bool"]
              },
              "size": {
                "type": "integer",
                "description": "返回结果数量，固定为 1",
                "enum": [1]
              },
              "sort": {
                "type": "array",
                "description": "排序规则数组，必须使用随机排序",
                "items": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "_script": {
                      "type": "object",
                      "additionalProperties": false,
                      "properties": {
                        "type": {"type": "string", "const": "number"},
                        "script": {
                          "type": "object",
                          "additionalProperties": false,
                          "properties": {
                            "source": {"type": "string", "const": "Math.random()"}
                          },
                          "required": ["source"]
                        },
                        "order": {"type": "string", "const": "asc"}
                      },
                      "required": ["type", "script", "order"]
                    }
                  },
                  "required": ["_script"]
                },
                "minItems": 1
              },
              "aggs": {
                "type": "object",
                "description": "聚合统计对象，必须包含remaining_questions",
                "additionalProperties": false,
                "properties": {
                  "remaining_questions": {
                    "type": "object",
                    "additionalProperties": false,
                    "properties": {
                      "cardinality": {
                        "type": "object",
                        "additionalProperties": false,
                        "properties": {
                          "field": {"type": "string", "const": "question_id"}
                        },
                        "required": ["field"]
                      }
                    },
                    "required": ["cardinality"]
                  }
                },
                "required": ["remaining_questions"]
              }
            },
            "required": ["query", "size", "sort", "aggs"]
          }
        },
        "required": ["index", "query"]
      }
    },
    "required": ["varName", "es_dsl"]
  }
}`,
			},
			{
				Name: "forward_result",
				Definition: `{
  "name": "forward_result",
  "description": "将结构化的JSON数据发送给前端用户。根据type字段区分数据类型：question(题目)、evaluation(评价)、summary(总结)",
  "strict": true,
  "parameters": {
    "type": "object",
    "properties": {
      "varName": {
        "type": "string",
        "description": "变量名，用于保存结果。题目用Question_N，评价用Evaluation_N，总结用Summary"
      },
      "result": {
        "description": "要发送的JSON对象，根据type字段匹配对应的结构",
        "anyOf": [
          {
            "type": "object",
            "description": "题目类型",
            "properties": {
              "type": {
                "type": "string",
                "const": "question",
                "description": "固定值question"
              },
              "question_id": {
                "type": "integer",
                "description": "题目ID，用于后续排除"
              },
              "question": {
                "type": "string",
                "description": "题目内容（不含答案）"
              },
              "remaining_questions": {
                "type": "integer",
                "description": "ES返回的剩余题数（包含当前题）"
              },
              "current": {
                "type": "integer",
                "description": "当前第几题（从1开始）"
              }
            },
            "required": ["type", "question_id", "question", "remaining_questions", "current"],
            "additionalProperties": false
          },
          {
            "type": "object",
            "description": "评价类型",
            "properties": {
              "type": {
                "type": "string",
                "const": "evaluation",
                "description": "固定值evaluation"
              },
              "question_id": {
                "type": "integer",
                "description": "题目ID，用于前端去重"
              },
              "scores": {
                "type": "object",
                "properties": {
                  "content_score": {
                    "type": "integer",
                    "minimum": 0,
                    "maximum": 100,
                    "description": "内容准确性得分"
                  },
                  "coverage_score": {
                    "type": "integer",
                    "minimum": 0,
                    "maximum": 100,
                    "description": "知识覆盖度得分"
                  },
                  "structure_score": {
                    "type": "integer",
                    "minimum": 0,
                    "maximum": 100,
                    "description": "表达清晰度得分"
                  }
                },
                "required": ["content_score", "coverage_score", "structure_score"],
                "additionalProperties": false
              },
              "evaluation": {
                "type": "object",
                "properties": {
                  "key_points_hit": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "答出的关键点列表"
                  },
                  "missed_points": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "遗漏的关键点列表"
                  },
                  "suggestion": {
                    "type": "string",
                    "description": "改进建议"
                  }
                },
                "required": ["key_points_hit", "missed_points", "suggestion"],
                "additionalProperties": false
              }
            },
            "required": ["type", "question_id", "scores", "evaluation"],
            "additionalProperties": false
          },
          {
            "type": "object",
            "description": "总结类型",
            "properties": {
              "type": {
                "type": "string",
                "const": "summary",
                "description": "固定值summary"
              },
              "total_questions": {
                "type": "integer",
                "description": "题库总题数"
              },
              "answered_questions": {
                "type": "integer",
                "description": "实际回答题数"
              },
              "overall_score": {
                "type": "integer",
                "minimum": 0,
                "maximum": 100,
                "description": "综合评分"
              },
              "strengths": {
                "type": "array",
                "items": {"type": "string"},
                "description": "优势点列表"
              },
              "weaknesses": {
                "type": "array",
                "items": {"type": "string"},
                "description": "薄弱点列表"
              },
              "priority_actions": {
                "type": "array",
                "items": {"type": "string"},
                "description": "优先改进建议列表"
              }
            },
            "required": ["type", "total_questions", "answered_questions", "overall_score", "strengths", "weaknesses", "priority_actions"],
            "additionalProperties": false
          }
        ]
      }
    },
    "required": ["varName", "result"],
    "additionalProperties": false
  }
}`,
			},
			{
				Name: "save_doc",
				Definition: `{
  "name": "save_doc",
  "description": "保存面试历史记录到变量中。用于保存单题的问答评价记录，或最终的总结报告。",
  "strict": true,
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
        "description": "JSON数组字符串。每个元素包含: question_id, question(题目), answer(回答), scores(评分对象), evaluation(评价对象)。必须包含之前的所有记录加上当前新记录。"
      }
    },
    "required": ["varName", "content", "type"],
    "additionalProperties": false
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

	// 5. 启动 gRPC 服务器
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
	bailianHdl *bailian.Handler,
	chatRepo *repository.ChatRepo,
	invConfigRepo *repository.InvocationConfigRepo,
	providerRepo *repository.ProviderRepository,
) stream.Handler {
	loadcfgHdl := loadcfg.NewLoadConfigHandler(chatRepo, invConfigRepo, providerRepo)
	renderHdl := render.NewHandler()
	storeHdl := store.NewHandler(chatRepo)

	// 组装责任链: loadcfg -> render -> store -> bailian
	loadcfgHdl.Next = renderHdl
	renderHdl.Next = storeHdl
	storeHdl.Next = bailianHdl
	bailianHdl.Handler = loadcfgHdl

	return loadcfgHdl
}

// TestInterviewProxyServer 启动 HTTP → gRPC 代理服务器
// 端口: 8080
// 功能: 将前端 HTTP 请求转换为 gRPC 调用，并将 gRPC 流式响应转换为 SSE
func TestInterviewProxyServer(t *testing.T) {
	// 0. 读取环境变量
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Fatal("未设置 API_KEY 环境变量")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		t.Fatal("未设置 BASE_URL 环境变量")
	}

	// COS 相关环境变量
	cosSecretID := os.Getenv("COS_SECRET_ID")
	if cosSecretID == "" {
		t.Fatal("未设置 COS_SECRET_ID 环境变量")
	}

	cosSecretKey := os.Getenv("COS_SECRET_KEY")
	if cosSecretKey == "" {
		t.Fatal("未设置 COS_SECRET_KEY 环境变量")
	}

	cosBucket := os.Getenv("COS_BUCKET_NAME")
	if cosBucket == "" {
		t.Fatal("未设置 COS_BUCKET_NAME 环境变量（格式：bucketname-appid）")
	}

	cosRegion := os.Getenv("COS_REGION")
	if cosRegion == "" {
		t.Fatal("未设置 COS_REGION 环境变量（如：ap-guangzhou）")
	}

	// 初始化 COS 客户端
	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cosBucket, cosRegion))
	b := &cos.BaseURL{BucketURL: u}
	cosClient := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cosSecretID,
			SecretKey: cosSecretKey,
		},
	})

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
		err = json.NewEncoder(w).Encode(map[string]string{
			"chat_sn": resp.Sn,
		})
		assert.NoError(t, err)
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

	// 5. 语音识别接口（上传到 COS + 百炼 qwen3-asr-flash）
	mux.HandleFunc("/api/audio/transcribe", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// 解析 multipart 表单
		err := r.ParseMultipartForm(32 << 20) // 最大 32MB
		if err != nil {
			http.Error(w, fmt.Sprintf("解析表单失败: %v", err), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, fmt.Sprintf("读取文件失败: %v", err), http.StatusBadRequest)
			return
		}
		defer file.Close()

		log.Printf("📤 收到音频文件: %s, 大小: %d 字节", header.Filename, header.Size)

		// 1. 上传到 COS（设置为公共读）
		// 生成唯一文件名：audio-temp/timestamp_filename.webm
		timestamp := time.Now().Unix()
		filename := fmt.Sprintf("audio-temp/%d_%s", timestamp, strings.Replace(header.Filename, " ", "_", -1))

		log.Printf("🔄 上传到 COS: %s", filename)

		// 设置上传选项（参考官方示例 Case2）
		opt := &cos.ObjectPutOptions{
			ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
				ContentType: "audio/webm",
			},
			ACLHeaderOptions: &cos.ACLHeaderOptions{
				XCosACL: "public-read", // 设置为公共读
			},
		}

		_, err = cosClient.Object.Put(context.Background(), filename, file, opt)
		if err != nil {
			log.Printf("❌ COS 上传失败: %v", err)
			http.Error(w, fmt.Sprintf("上传文件失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 2. 构建 COS 文件的公网访问 URL
		cosFileURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", cosBucket, cosRegion, filename)
		log.Printf("✅ COS 上传成功: %s", cosFileURL)

		// 3. 调用百炼语音识别 API（异步）
		// 使用 qwen3-asr-flash：无需申请，直接可用
		log.Printf("🔄 调用百炼 qwen3-asr-flash 识别（异步）...")
		transcribeReq := map[string]any{
			"model": "qwen3-asr-flash", // 支持中文、英文，无需申请即可使用
			"input": map[string]any{
				"messages": []map[string]any{
					{
						"role": "system",
						"content": []map[string]any{
							{
								"text": "请将以下音频文件URL转换为文字：",
							},
						},
					},
					{
						"role": "user",
						"content": []map[string]any{
							{
								"audio": cosFileURL,
							},
						},
					},
				},
			},
			"parameters": map[string]any{
				"asr_options": map[string]any{
					"enable_itn": true,
				},
			},
		}

		reqBody, _ := json.Marshal(transcribeReq)

		// 构建完整URL（注意：baseURL 可能已经包含 /api/v1）
		// 如果 baseURL 已经包含 /api/v1，则只拼接 /services/audio/asr/transcription
		// 否则拼接完整路径 /api/v1/services/audio/asr/transcription
		var fullURL string
		if strings.HasSuffix(baseURL, "/api/v1") {
			fullURL = baseURL + "/services/aigc/multimodal-generation/generation"
		} else {
			fullURL = baseURL + "/api/v1/services/aigc/multimodal-generation/generation"
		}

		log.Printf("📡 请求URL: %s", fullURL)
		log.Printf("📦 请求体: %s", string(reqBody))
		log.Printf("🔑 API Key (前10字符): %s...", apiKey[:min(10, len(apiKey))])

		httpReq, err := http.NewRequest("POST", fullURL, bytes.NewReader(reqBody))
		if err != nil {
			log.Printf("❌ 创建请求失败: %v", err)
			http.Error(w, fmt.Sprintf("创建请求失败: %v", err), http.StatusInternalServerError)
			return
		}

		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
		httpReq.Header.Set("Content-Type", "application/json")

		httpClient := &http.Client{Timeout: 60 * time.Second}
		resp, err := httpClient.Do(httpReq)
		if err != nil {
			log.Printf("❌ 识别请求失败: %v", err)
			http.Error(w, fmt.Sprintf("识别请求失败: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("❌ 读取响应失败: %v", err)
			http.Error(w, fmt.Sprintf("读取响应失败: %v", err), http.StatusInternalServerError)
			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("❌ 提交任务失败: HTTP %d, 响应: %s", resp.StatusCode, string(respBody))
			http.Error(w, fmt.Sprintf("提交任务失败: %s", string(respBody)), resp.StatusCode)
			return
		}

		// 4. 解析文本
		var submitResp struct {
			Output struct {
				Choices []struct {
					Message struct {
						Content []struct {
							Text string `json:"text"`
						} `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			} `json:"output"`
		}

		err = json.Unmarshal(respBody, &submitResp)
		if err != nil || len(submitResp.Output.Choices) == 0 || len(submitResp.Output.Choices[0].Message.Content) == 0 {
			log.Printf("❌ 解析任务ID失败: %v, 原始响应: %s", err, string(respBody))
			http.Error(w, "解析任务ID失败", http.StatusInternalServerError)
			return
		}

		transcribedText := submitResp.Output.Choices[0].Message.Content[0].Text
		log.Printf("✅ 识别成功: %s", transcribedText)

		// 5. 返回结果给前端
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{
			"text":    transcribedText,
			"cos_url": cosFileURL,
		})
		assert.NoError(t, err)
	}))

	// 6. 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"grpc":   "localhost:9090",
			"time":   time.Now().Format(time.RFC3339),
		})
		assert.NoError(t, err)
	})

	log.Println("🚀 HTTP 代理服务器启动于 :8080")
	log.Println("📡 转发目标: localhost:9090 (gRPC)")
	log.Println("📍 端点:")
	log.Println("   - POST /api/interview/chat/create  (创建会话)")
	log.Println("   - POST /api/interview/stream        (流式面试)")
	log.Println("   - POST /api/audio/transcribe        (语音识别: COS + qwen3-asr-flash)")
	log.Println("   - GET  /health                      (健康检查)")
	log.Println("💡 按 Ctrl+C 停止服务器")
	log.Println("---")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		t.Fatalf("HTTP 服务器启动失败: %v", err)
	}
}

// truncate 截断字符串用于日志显示
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// prepareKbaseTestData 准备Kbase测试数据
// 使用 go-elasticsearch 直接操作ES，创建索引并插入测试题目
func prepareKbaseTestData(t *testing.T) {
	// 1. 创建ES客户端
	esAddr := "http://localhost:9200"
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esAddr},
	})
	require.NoError(t, err, "创建ES客户端失败")

	// 2. 健康检查（最多等待5秒）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = esClient.Ping(esClient.Ping.WithContext(ctx))
	require.NoError(t, err, "ES不可用（等待5秒超时），请先启动Elasticsearch服务")

	indexName := "interview_questions_mysql"

	// 3. 删除已存在的索引（如果有）
	_, err = esClient.Indices.Delete([]string{indexName})
	assert.NoError(t, err)

	// 4. 创建索引和Mapping
	createIndexWithMapping(t, esClient, indexName)

	// 5. 插入测试数据
	questions := buildTestQuestions()
	for _, q := range questions {
		insertQuestion(t, esClient, indexName, q)
	}

	// 6. 刷新索引，确保数据可搜索
	_, err = esClient.Indices.Refresh(esClient.Indices.Refresh.WithIndex(indexName))
	require.NoError(t, err)
}

// createIndexWithMapping 创建索引和Mapping
func createIndexWithMapping(t *testing.T, client *elasticsearch.Client, indexName string) {
	mapping := `{
  "mappings": {
    "properties": {
      "level": { "type": "keyword" },
      "question_id": { "type": "integer" },
      "title": { "type": "text" },
      "analysis": { "type": "text" },
      "tags": { "type": "keyword" },
      "answers": { "type": "object", "enabled": false },
      "created_at": { "type": "date" },
      "updated_at": { "type": "date" }
    }
  }
}`

	resp, err := client.Indices.Create(
		indexName,
		client.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	require.NoError(t, err, "创建索引失败")
	defer resp.Body.Close()

	if resp.IsError() {
		body, _ := io.ReadAll(resp.Body)
		require.Failf(t, "创建索引失败", "响应: %s", string(body))
	}
}

// buildTestQuestions 构建测试题目数据
func buildTestQuestions() []map[string]any {
	now := time.Now().Format(time.RFC3339)

	return []map[string]any{
		{
			"question_id": 1,
			"level":       "junior",
			"title":       "什么是事务？请简述ACID特性",
			"analysis":    "事务是数据库操作的基本单位，ACID是事务的四个基本特性：原子性(Atomicity)、一致性(Consistency)、隔离性(Isolation)、持久性(Durability)。",
			"tags":        []string{"事务", "ACID", "基础概念"},
			"answers": map[string]any{
				"15k": map[string]any{
					"content":    "事务是保证一组数据库操作要么全部成功，要么全部失败的机制。ACID是：原子性(Atomicity)表示事务不可分割、一致性(Consistency)保证数据完整性、隔离性(Isolation)多个事务互不干扰、持久性(Durability)事务提交后永久保存。",
					"key_points": []string{"原子性", "一致性", "隔离性", "持久性"},
				},
				"25k": map[string]any{
					"content":    "在15K基础上，需要理解四种隔离级别：读未提交(Read Uncommitted)、读已提交(Read Committed)、可重复读(Repeatable Read)、串行化(Serializable)。MySQL默认使用可重复读，并通过MVCC(多版本并发控制)实现。",
					"key_points": []string{"隔离级别", "MVCC", "可重复读", "幻读"},
				},
				"35k": map[string]any{
					"content":    "在25K基础上，还需深入：锁机制(行锁、表锁、间隙锁、Next-Key Lock)、事务日志(undo log用于回滚、redo log用于持久化)、两阶段提交协议、事务优化实践(减小事务范围、避免长事务)。",
					"key_points": []string{"锁机制", "undo/redo log", "两阶段提交", "性能优化"},
				},
			},
			"created_at": now,
			"updated_at": now,
		},
		{
			"question_id": 2,
			"level":       "junior",
			"title":       "什么是索引？索引的作用是什么？",
			"analysis":    "索引是帮助MySQL高效获取数据的数据结构，类似于书的目录。MySQL主要使用B+树作为索引结构。",
			"tags":        []string{"索引", "B+树", "查询优化"},
			"answers": map[string]any{
				"15k": map[string]any{
					"content":    "索引是一种数据结构，可以加快数据查询速度。就像书的目录，可以快速找到需要的内容，而不用逐页翻阅。MySQL使用B+树作为索引结构，通过空间换时间的方式提升查询性能。",
					"key_points": []string{"B+树", "查询加速", "空间换时间"},
				},
				"25k": map[string]any{
					"content":    "需要补充索引类型：聚簇索引(主键索引，数据和索引在一起)、非聚簇索引(二级索引，需要回表)、覆盖索引(查询列都在索引中，不需要回表)。理解索引的优缺点：优点是加速查询，缺点是占用空间且降低写入性能。",
					"key_points": []string{"聚簇索引", "非聚簇索引", "覆盖索引", "回表"},
				},
				"35k": map[string]any{
					"content":    "深入理解索引失效场景(函数操作、类型转换、like左模糊)、最左前缀原则(联合索引只能从最左边开始使用)、索引下推(ICP)、索引优化实践(选择性高的列、避免冗余索引)。",
					"key_points": []string{"索引失效", "最左前缀", "索引下推", "索引优化"},
				},
			},
			"created_at": now,
			"updated_at": now,
		},
		{
			"question_id": 3,
			"level":       "middle", // 干扰项
			"title":       "如何排查慢查询？",
			"analysis":    "慢查询优化是数据库性能调优的重要环节，需要从日志分析、执行计划、索引优化等多个角度入手。",
			"tags":        []string{"慢查询", "性能优化", "explain"},
			"answers": map[string]any{
				"15k": map[string]any{
					"content":    "使用slow_query_log查看慢查询日志，找出执行时间超过阈值的SQL语句。",
					"key_points": []string{"slow_query_log", "执行时间"},
				},
				"25k": map[string]any{
					"content":    "使用EXPLAIN分析查询计划，重点关注type、key、rows等字段。type最好是ref或const，避免ALL全表扫描。",
					"key_points": []string{"EXPLAIN", "索引使用", "type字段"},
				},
				"35k": map[string]any{
					"content":    "深入分析执行计划、优化SQL(子查询改JOIN、避免SELECT *)、添加合适的索引、考虑分库分表、使用缓存。",
					"key_points": []string{"执行计划优化", "索引设计", "SQL重写", "分库分表"},
				},
			},
			"created_at": now,
			"updated_at": now,
		},
		{
			"question_id": 4,
			"level":       "junior",
			"title":       "请解释COUNT(*)、COUNT(1)和COUNT(column)的区别",
			"analysis":    "COUNT是常用的聚合函数，用于统计行数，但不同的写法有不同的含义。",
			"tags":        []string{"COUNT", "聚合函数", "SQL"},
			"answers": map[string]any{
				"15k": map[string]any{
					"content":    "COUNT(*)统计所有行（包括NULL），COUNT(column)统计该列非NULL的行数，COUNT(1)和COUNT(*)效果相同。",
					"key_points": []string{"COUNT(*)", "NULL处理", "行数统计"},
				},
				"25k": map[string]any{
					"content":    "COUNT(1)和COUNT(*)性能基本相同，MySQL优化器会自动优化。COUNT(column)需要判断NULL，性能略低。",
					"key_points": []string{"性能对比", "优化器", "NULL判断"},
				},
				"35k": map[string]any{
					"content":    "不同存储引擎的COUNT实现差异：MyISAM保存了表的行数，COUNT(*)很快；InnoDB需要扫描，可以通过添加索引或使用缓存优化。",
					"key_points": []string{"InnoDB", "MyISAM", "实现原理", "优化方案"},
				},
			},
			"created_at": now,
			"updated_at": now,
		},
		//{
		//	"question_id": 5,
		//	"level":       "junior",
		//	"title":       "数据库设计的三大范式是什么？",
		//	"analysis":    "数据库范式是设计关系数据库的基本原则，用于减少数据冗余和提高数据完整性。",
		//	"tags":        []string{"范式", "数据库设计", "规范化"},
		//	"answers": map[string]any{
		//		"15k": map[string]any{
		//			"content":    "第一范式(1NF)：列不可再分，每个字段都是原子性的。第二范式(2NF)：消除部分依赖，非主键列完全依赖于主键。第三范式(3NF)：消除传递依赖，非主键列不依赖于其他非主键列。",
		//			"key_points": []string{"1NF", "2NF", "3NF", "原子性"},
		//		},
		//		"25k": map[string]any{
		//			"content":    "需要举例说明：如订单表包含客户信息违反2NF，应拆分为订单表和客户表。理解反范式化：为了性能有时会适当冗余数据。",
		//			"key_points": []string{"实际案例", "表拆分", "反范式化"},
		//		},
		//		"35k": map[string]any{
		//			"content":    "理解BCNF(消除主属性对码的部分和传递依赖)、4NF(消除多值依赖)。掌握反范式化的应用场景：高并发读场景、数据仓库、适当的冗余可以减少JOIN提升性能。",
		//			"key_points": []string{"BCNF", "4NF", "反范式化场景", "性能权衡"},
		//		},
		//	},
		//	"created_at": now,
		//	"updated_at": now,
		//},
		//{
		//	"question_id": 6,
		//	"level":       "junior",
		//	"title":       "请解释MySQL的主键和外键",
		//	"analysis":    "主键和外键是关系数据库的核心概念，用于唯一标识和建立表之间的关联。",
		//	"tags":        []string{"主键", "外键", "约束"},
		//	"answers": map[string]any{
		//		"15k": map[string]any{
		//			"content":    "主键(Primary Key)唯一标识表中的每一条记录，不能为NULL。外键(Foreign Key)用于建立表之间的关联关系，外键的值必须是另一个表的主键值或NULL。",
		//			"key_points": []string{"主键唯一性", "外键约束", "表关联"},
		//		},
		//		"25k": map[string]any{
		//			"content":    "主键选择策略：自增ID(简单高效)、UUID(全局唯一但无序)、业务主键(有业务含义)。外键的级联操作：ON DELETE CASCADE(级联删除)、ON UPDATE CASCADE(级联更新)。",
		//			"key_points": []string{"主键选择", "级联删除", "级联更新"},
		//		},
		//		"35k": map[string]any{
		//			"content":    "分布式场景下的主键设计：雪花算法(Snowflake)生成分布式ID，兼顾唯一性和趋势递增。外键的性能影响：外键会降低写入性能，高并发场景通常在应用层维护关联关系而不使用外键约束。",
		//			"key_points": []string{"分布式ID", "雪花算法", "外键性能", "应用层约束"},
		//		},
		//	},
		//	"created_at": now,
		//	"updated_at": now,
		//},
	}
}

// insertQuestion 插入单个题目到ES
func insertQuestion(t *testing.T, client *elasticsearch.Client, indexName string, doc map[string]any) {
	docJSON, err := json.Marshal(doc)
	require.NoError(t, err, "序列化文档失败")

	docID := fmt.Sprintf("mysql_%03d", doc["question_id"])

	resp, err := client.Index(
		indexName,
		bytes.NewReader(docJSON),
		client.Index.WithDocumentID(docID),
	)
	require.NoError(t, err, "插入文档失败")
	defer resp.Body.Close()

	if resp.IsError() {
		body, _ := io.ReadAll(resp.Body)
		require.Failf(t, "插入文档失败", "[%s]: %s", resp.Status(), string(body))
	}
}

// TestKbaseQuestionQuery 测试Kbase题库查询场景
func TestKbaseQuestionQuery(t *testing.T) {
	// 前置准备：创建测试数据
	prepareKbaseTestData(t)

	kbaseBaseURL := "http://localhost:8082"
	indexName := "interview_questions_mysql"

	// 表格驱动测试
	tests := []struct {
		name      string
		query     map[string]any
		validator func(t *testing.T, result map[string]any)
	}{
		{
			name:  "场景1_获取第一题",
			query: buildFirstQuestionQuery(),
			validator: func(t *testing.T, result map[string]any) {
				// 验证返回1个结果
				hits := result["hits"].(map[string]any)["hits"].([]any)
				assert.Equal(t, 1, len(hits), "应该返回1个结果")

				// 验证是第一题（question_id=1）
				firstHit := hits[0].(map[string]any)
				source := firstHit["_source"].(map[string]any)
				qid := int(source["question_id"].(float64))
				assert.Equal(t, 1, qid, "应该返回第一题")

				// ES返回的包含当前题，所以是5（题目1,2,4,5,6）
				aggs := result["aggregations"].(map[string]any)
				remaining := aggs["remaining_questions"].(map[string]any)["value"].(float64)
				assert.Equal(t, float64(5), remaining, "ES应该返回5（包含当前题）")
			},
		},
		{
			name:  "场景2_排除2道题获取下一题",
			query: buildNextQuestionQuery([]int{1, 2}), // 排除1和2
			validator: func(t *testing.T, result map[string]any) {
				hits := result["hits"].(map[string]any)["hits"].([]any)
				assert.Equal(t, 1, len(hits), "应该返回1个结果")

				// 验证返回的不是1或2
				source := hits[0].(map[string]any)["_source"].(map[string]any)
				qid := int(source["question_id"].(float64))
				assert.NotContains(t, []int{1, 2}, qid, "不应该返回已问过的题目")
				assert.Contains(t, []int{4, 5, 6}, qid, "应该返回4、5、6之一")

				// ES返回的包含当前题，所以是3（题目4,5,6）
				aggs := result["aggregations"].(map[string]any)
				remaining := aggs["remaining_questions"].(map[string]any)["value"].(float64)
				assert.Equal(t, float64(3), remaining, "ES应该返回3（包含当前题）")
			},
		},
		{
			name:  "场景3_排除4道题获取最后一题",
			query: buildNextQuestionQuery([]int{1, 2, 4, 5}), // 排除4道，只剩题目6
			validator: func(t *testing.T, result map[string]any) {
				hits := result["hits"].(map[string]any)["hits"].([]any)
				assert.Equal(t, 1, len(hits), "应该返回1个结果")

				// 验证返回的是题目6
				source := hits[0].(map[string]any)["_source"].(map[string]any)
				qid := int(source["question_id"].(float64))
				assert.Equal(t, 6, qid, "应该返回最后一题(ID=6)")

				// ES返回的包含当前题，所以是1（只有题目6）
				aggs := result["aggregations"].(map[string]any)
				remaining := aggs["remaining_questions"].(map[string]any)["value"].(float64)
				assert.Equal(t, float64(1), remaining, "ES应该返回1（只剩最后一题）")
			},
		},
		{
			name:  "场景4_排除所有题目后返回空",
			query: buildNextQuestionQuery([]int{1, 2, 4, 5, 6}), // 排除所有junior题目
			validator: func(t *testing.T, result map[string]any) {
				hits := result["hits"].(map[string]any)["hits"].([]any)
				assert.Equal(t, 0, len(hits), "排除所有题目后应该返回空")

				// 没有题了，剩余数=0
				aggs := result["aggregations"].(map[string]any)
				remaining := aggs["remaining_questions"].(map[string]any)["value"].(float64)
				assert.Equal(t, float64(0), remaining, "排除所有题目后剩余数应为0")
			},
		},
		{
			name:  "场景5_验证level过滤有效",
			query: buildFirstQuestionQuery(), // 只查junior
			validator: func(t *testing.T, result map[string]any) {
				hits := result["hits"].(map[string]any)["hits"].([]any)
				require.Greater(t, len(hits), 0, "应该有返回结果")

				source := hits[0].(map[string]any)["_source"].(map[string]any)

				// 验证level是junior
				assert.Equal(t, "junior", source["level"], "level应该是junior")

				// 验证question_id不是3（3是middle级别）
				qid := int(source["question_id"].(float64))
				assert.NotEqual(t, 3, qid, "不应该返回middle级别的题目(ID=3)")
			},
		},
	}

	// 执行所有测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executeESQuery(t, kbaseBaseURL, indexName, tt.query)
			tt.validator(t, result)
		})
	}
}

// buildFirstQuestionQuery 构建获取第一题的查询
func buildFirstQuestionQuery() map[string]any {
	return map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{"term": map[string]any{"level": "junior"}},
				},
			},
		},
		"size": 1,
		"sort": []map[string]any{
			{"question_id": map[string]any{"order": "asc"}},
		},
		"aggs": map[string]any{
			"remaining_questions": map[string]any{
				"cardinality": map[string]any{
					"field": "question_id",
				},
			},
		},
	}
}

// buildNextQuestionQuery 构建获取下一题的查询（排除已问）
func buildNextQuestionQuery(excludeIDs []int) map[string]any {
	return map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{"term": map[string]any{"level": "junior"}},
				},
				"must_not": []map[string]any{
					{"terms": map[string]any{"question_id": excludeIDs}},
				},
			},
		},
		"size": 1,
		"sort": []map[string]any{
			{
				"_script": map[string]any{
					"type": "number",
					"script": map[string]any{
						"source": "Math.random()",
					},
					"order": "asc",
				},
			},
		},
		"aggs": map[string]any{
			"remaining_questions": map[string]any{
				"cardinality": map[string]any{
					"field": "question_id",
				},
			},
		},
	}
}

// executeESQuery 执行ES查询
func executeESQuery(t *testing.T, baseURL, index string, query map[string]any) map[string]any {
	// 调用 kbase 的 /api/v1/es_search 接口
	requestBody := map[string]any{
		"index": index,
		"query": query,
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err, "序列化请求失败")

	resp, err := http.Post(
		baseURL+"/api/v1/es_search",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	require.NoError(t, err, "HTTP请求失败")
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode, "请求应该成功")

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "解析响应失败")

	return result
}
