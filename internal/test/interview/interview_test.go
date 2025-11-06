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

package interview

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
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	chatv1 "github.com/ecodeclub/ai-gateway-go/api/proto/gen/chat/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	igrpc "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/orchestrator"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall/forward"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall/kbase"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall/multifunc"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/fcall/savedoc"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/loadcfg"
	openaistream "github.com/ecodeclub/ai-gateway-go/internal/service/stream/openai"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/render"
	"github.com/ecodeclub/ai-gateway-go/internal/service/stream/store"
	_ "github.com/ecodeclub/ai-gateway-go/internal/test"
	testioc "github.com/ecodeclub/ai-gateway-go/internal/test/ioc"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/gotomicro/ego/core/elog"
	openai3 "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 新的系统提示词文件
//
//go:embed system_prompt_get_question.md
var systemPromptGetQuestion string

//go:embed system_prompt_evaluate_save.md
var systemPromptEvaluateSave string

//go:embed system_prompt_summary_save.md
var systemPromptSummarySave string

//go:embed system_prompt_send.md
var systemPromptSend string

// 新的用户提示词文件
//
//go:embed user_prompt_get_question.md
var userPromptGetQuestion string

//go:embed user_prompt_evaluate_save.md
var userPromptEvaluateSave string

//go:embed user_prompt_summary_save.md
var userPromptSummarySave string

//go:embed user_prompt_send.md
var userPromptSend string

// TestGrpcServer 启动 gRPC 服务器用于面试功能测试
// 端口: 9090
// 功能: 提供 Stream 接口，处理面试逻辑
func TestGrpcServer(t *testing.T) {
	elog.DefaultLogger.SetLevel(elog.DebugLevel)
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

	// 注册 multi_call function call（需要 Registry，放在最后）
	multiCallFCall := multifunc.NewFCall()
	multiCallFCall.Registry = registry
	registry.Register(multiCallFCall)

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
	orchestratorHdl := orchestrator.NewOrchestrator(streamHandler)

	app := testioc.InitApp(testioc.TestOnly{
		Orchestrator: orchestratorHdl,
	})

	// 4. 准备测试数据
	log.Println("准备测试数据...")

	// 4.1 创建 Provider（OpenAI）
	providerID, err := providerDAO.SaveProvider(ctx, dao.Provider{
		Name:   "OpenAI",
		APIKey: "test-api-key", // 测试环境使用占位符
	})
	if err != nil {
		t.Fatalf("创建 Provider 失败: %v", err)
	}
	log.Printf("   ✓ 创建 Provider: OpenAI (ID: %d)", providerID)

	// 4.2 创建 Model
	modelID, err := providerDAO.SaveModel(ctx, dao.Model{
		Name:        openai3.ChatModelGPT5,
		Pid:         providerID,
		InputPrice:  150, // $0.150 / 1M tokens
		OutputPrice: 600, // $0.600 / 1M tokens
		PriceMode:   "token",
	})
	if err != nil {
		t.Fatalf("创建 Model 失败: %v", err)
	}
	log.Printf("   ✓ 创建 Model: gpt-5 (ID: %d)", modelID)

	// 4.3 创建 Biz
	bizConfigDAO := dao.NewBizConfigDAO(db)
	bizRepo := repository.NewBizConfigRepository(bizConfigDAO)
	bizSvc := service.NewBizConfigService(bizRepo)
	biz := domain.Biz{
		ID:        100000,
		Name:      "MySQL模拟面试助手",
		OwnerID:   1,
		OwnerType: "user",
	}
	_, err = bizSvc.Save(ctx, biz)
	require.NoError(t, err, "创建 Biz 失败")
	log.Printf("创建 Biz 成功，ID = %d", biz.ID)

	// 4.4 创建 4 个 InvocationConfig
	invSvc := service.NewInvocationConfigService(invConfigRepo, bizRepo, providerRepo)

	// 定义 4 个 ConfigID
	cfgIDGetQuestion := int64(100002)  // 获取题目
	cfgIDEvaluateSave := int64(100003) // 评价答案并保存历史
	cfgIDSummarySave := int64(100004)  // 生成总结并保存
	cfgIDSend := int64(100005)         // 发送题目

	// 创建 get_question 配置
	_, err = invSvc.Save(ctx, domain.InvocationConfig{
		ID:          cfgIDGetQuestion,
		Name:        "获取题目",
		Biz:         domain.Biz{ID: biz.ID},
		Description: "获取面试题目（第一题或下一题）",
	})
	require.NoError(t, err)
	log.Printf("创建 InvocationConfig: 获取题目 (ID: %d)", cfgIDGetQuestion)

	// 创建 evaluate_and_save 配置
	_, err = invSvc.Save(ctx, domain.InvocationConfig{
		ID:          cfgIDEvaluateSave,
		Name:        "评价答案并保存历史",
		Biz:         domain.Biz{ID: biz.ID},
		Description: "评价用户答案并保存历史记录",
	})
	require.NoError(t, err)
	log.Printf("创建 InvocationConfig: 评价答案并保存历史 (ID: %d)", cfgIDEvaluateSave)

	// 创建 summary_and_save 配置
	_, err = invSvc.Save(ctx, domain.InvocationConfig{
		ID:          cfgIDSummarySave,
		Name:        "生成总结并保存",
		Biz:         domain.Biz{ID: biz.ID},
		Description: "生成面试总结并保存",
	})
	require.NoError(t, err)
	log.Printf("创建 InvocationConfig: 生成总结并保存 (ID: %d)", cfgIDSummarySave)

	// 创建 send_to_user 配置
	_, err = invSvc.Save(ctx, domain.InvocationConfig{
		ID:          cfgIDSend,
		Name:        "发送题目给用户",
		Biz:         domain.Biz{ID: biz.ID},
		Description: "将题目格式化后发送给前端用户",
	})
	require.NoError(t, err)
	log.Printf("创建 InvocationConfig: 发送题目给用户 (ID: %d)", cfgIDSend)

	// 更新 Biz 设置 Orchestration
	biz.Config = domain.BizConfig{
		Orchestration: domain.Orchestration{
			Main: nil, // Main 不再使用，前端直接传递 state 参数
			Threads: map[string]*domain.Thread{
				"get_question":      {CfgID: cfgIDGetQuestion},
				"evaluate_and_save": {CfgID: cfgIDEvaluateSave},
				"summary_and_save":  {CfgID: cfgIDSummarySave},
				"send_to_user":      {CfgID: cfgIDSend},
			},
		},
	}
	_, err = bizSvc.Save(ctx, biz)
	require.NoError(t, err, "更新 Biz Orchestration 失败")
	log.Printf("更新 Biz Orchestration (Threads: get_question=%d, evaluate_and_save=%d, summary_and_save=%d, send_to_user=%d)",
		cfgIDGetQuestion, cfgIDEvaluateSave, cfgIDSummarySave, cfgIDSend)

	// 4.5 创建 4 个 InvocationConfigVersion（active）
	// 由于代码较长，我会创建一个辅助函数来生成函数定义JSON
	// 创建 get_question 的 Version
	versionIDGetQuestion, err := invSvc.SaveVersion(ctx, domain.InvocationConfigVersion{
		Config:       domain.InvocationConfig{ID: cfgIDGetQuestion},
		Model:        domain.Model{ID: modelID},
		Version:      "v1.0",
		Status:       domain.InvocationCfgVersionStatusActive,
		SystemPrompt: systemPromptGetQuestion,
		Prompt:       userPromptGetQuestion,
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
      "nextState": {
        "type": "string",
        "enum": ["send_to_user", ""],
        "description": "下一个状态。必须从枚举值中选择。固定为 send_to_user。"
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
    "required": ["varName", "es_dsl", "nextState"]
  }
}`,
			},
		},
	})
	require.NoError(t, err, "创建 get_question Version 失败")
	log.Printf("   ✓ 创建 InvocationConfigVersion: get_question v1.0 (ID: %d)", versionIDGetQuestion)

	// 创建 evaluate_and_save 的 Version（使用 multi_call）
	versionIDEvaluateSave, err := invSvc.SaveVersion(ctx, domain.InvocationConfigVersion{
		Config:       domain.InvocationConfig{ID: cfgIDEvaluateSave},
		Model:        domain.Model{ID: modelID},
		Version:      "v1.0",
		Status:       domain.InvocationCfgVersionStatusActive,
		SystemPrompt: systemPromptEvaluateSave,
		Prompt:       userPromptEvaluateSave,
		Temperature:  0,
		TopP:         1.0,
		MaxTokens:    200000,
		Functions: []domain.Function{
			{
				Name:       "multi_call",
				Definition: getMultiCallFunctionDefinition(),
			},
		},
	})
	require.NoError(t, err, "创建 evaluate_and_save Version 失败")
	log.Printf("   ✓ 创建 InvocationConfigVersion: evaluate_and_save v1.0 (ID: %d)", versionIDEvaluateSave)

	// 创建 summary_and_save 的 Version（使用 multi_call）
	versionIDSummarySave, err := invSvc.SaveVersion(ctx, domain.InvocationConfigVersion{
		Config:       domain.InvocationConfig{ID: cfgIDSummarySave},
		Model:        domain.Model{ID: modelID},
		Version:      "v1.0",
		Status:       domain.InvocationCfgVersionStatusActive,
		SystemPrompt: systemPromptSummarySave,
		Prompt:       userPromptSummarySave,
		Temperature:  0,
		TopP:         1.0,
		MaxTokens:    200000,
		Functions: []domain.Function{
			{
				Name:       "multi_call",
				Definition: getMultiCallFunctionDefinition(),
			},
		},
	})
	require.NoError(t, err, "创建 summary_and_save Version 失败")
	log.Printf("   ✓ 创建 InvocationConfigVersion: summary_and_save v1.0 (ID: %d)", versionIDSummarySave)

	// 创建 send_to_user 的 Version（使用 forward_result）
	versionIDSend, err := invSvc.SaveVersion(ctx, domain.InvocationConfigVersion{
		Config:       domain.InvocationConfig{ID: cfgIDSend},
		Model:        domain.Model{ID: modelID},
		Version:      "v1.0",
		Status:       domain.InvocationCfgVersionStatusActive,
		SystemPrompt: systemPromptSend,
		Prompt:       userPromptSend,
		Temperature:  0,
		TopP:         1.0,
		MaxTokens:    200000,
		Functions: []domain.Function{
			{
				Name: "forward_result",
				Definition: `{
  "name": "forward_result",
  "description": "将结构化的JSON数据发送给前端用户。根据type字段区分数据类型：question(题目)、evaluation(评价)、summary(总结)",
  "strict": true,
  "parameters": {
    "type": "object",
    "additionalProperties": false,
    "properties": {
      "varName": {
        "type": "string",
        "description": "变量名，用于保存结果。题目用Question_N，评价用Evaluation_N，总结用Summary"
      },
      "nextState": {
        "type": "string",
        "enum": [""],
        "description": "下一个状态。固定为空字符串，表示结束流程。"
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
    "required": ["varName", "result", "nextState"],
    "additionalProperties": false
  }
}`,
			},
		},
	})
	require.NoError(t, err, "创建 send_to_user Version 失败")
	log.Printf("   ✓ 创建 InvocationConfigVersion: send_to_user v1.0 (ID: %d)", versionIDSend)

	// 清理函数 - 使用 t.Cleanup 确保在测试失败、panic 或正常结束时都会执行
	// 注意：如果进程被强制终止（如 IDE 中点击停止、Ctrl+C、kill -9），t.Cleanup 也不会执行
	t.Cleanup(func() {
		log.Println("\n清理测试数据...")
		// 清理 Chat 相关数据（先清理依赖数据）
		// 先查询测试用户的所有 Chat，然后删除相关的 Turn
		var chats []dao.Chat
		db.Where("uid = ?", 123).Find(&chats)
		for _, chat := range chats {
			db.Where("chat_sn = ?", chat.Sn).Delete(&dao.Turn{})
		}
		db.Where("uid = ?", 123).Delete(&dao.Chat{}) // 清理测试用户的 Chat
		// 清理配置数据
		db.Delete(&dao.InvocationConfigVersion{}, versionIDSend)
		db.Delete(&dao.InvocationConfigVersion{}, versionIDSummarySave)
		db.Delete(&dao.InvocationConfigVersion{}, versionIDEvaluateSave)
		db.Delete(&dao.InvocationConfigVersion{}, versionIDGetQuestion)
		db.Delete(&dao.InvocationConfig{}, cfgIDSend)
		db.Delete(&dao.InvocationConfig{}, cfgIDSummarySave)
		db.Delete(&dao.InvocationConfig{}, cfgIDEvaluateSave)
		db.Delete(&dao.InvocationConfig{}, cfgIDGetQuestion)
		db.Delete(&dao.Biz{}, biz.ID)
		db.Delete(&dao.Model{}, modelID)
		db.Delete(&dao.Provider{}, providerID)
		log.Println("测试数据已清理")
	})

	log.Println("\n数据准备完成，测试环境已就绪")

	// 5. 启动 gRPC 服务器
	chatSvc := app.ChatService
	chatServer := igrpc.NewChatServer(chatSvc, orchestratorHdl, bizSvc)

	grpcServer := grpc.NewServer()
	chatv1.RegisterServiceServer(grpcServer, chatServer)

	lis, err := net.Listen("tcp", "localhost:9090")
	if err != nil {
		t.Fatalf("监听端口失败: %v", err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC 服务器错误: %v", err)
		}
	}()

	// 优雅关闭：监听信号并执行清理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("gRPC 服务器启动于 localhost:9090")
	log.Printf("Biz.ID: %d", biz.ID)
	log.Println("按 Ctrl+C 停止服务器")
	log.Println("---")

	// 等待终止信号
	<-sigChan
	log.Println("\n收到终止信号，正在优雅关闭...")

	// 停止 gRPC 服务器
	grpcServer.Stop()
	log.Println("gRPC 服务器已停止")

	// 注意：t.Cleanup 会在测试函数返回时自动执行，这里不需要手动调用
	// 但为了确保清理，我们也可以手动触发清理逻辑（如果需要立即清理）
	// 实际上，让测试正常返回，t.Cleanup 会自动执行
}

// getMultiCallFunctionDefinition 返回 multi_call 函数的 JSON Schema 定义
func getMultiCallFunctionDefinition() string {
	return `{
  "name": "multi_call",
  "description": "依次执行多个函数调用。所有函数调用会按顺序执行，最后一个函数的 nextState 会作为整个 multi_call 的 nextState。",
  "strict": true,
  "parameters": {
    "type": "object",
    "additionalProperties": false,
    "properties": {
      "content": {
        "type": "string",
        "description": "返回给LLM的确认信息。用于表示所有函数调用已执行完成，可以是总结性的确认文本，如 \"所有函数调用已执行完成\" 或 \"操作已完成\"。"
      },
      "calls": {
        "type": "array",
        "description": "要执行的函数调用列表，按顺序执行",
        "items": {
          "type": "object",
          "additionalProperties": false,
          "properties": {
            "name": {
              "type": "string",
              "enum": ["forward_result", "save_doc"],
              "description": "函数名。必须是 forward_result 或 save_doc 之一。"
            },
            "arguments": {
              "description": "函数参数对象。根据 name 字段的值，构造对应的参数结构。",
              "anyOf": [
                {
                  "type": "object",
                  "description": "forward_result 的参数结构",
                  "properties": {
                    "varName": {
                      "type": "string",
                      "description": "变量名，用于保存结果。题目用QuestionOutput_N，评价用EvaluationOutput_N，总结用SummaryOutput"
                    },
                    "nextState": {
                      "type": "string",
                      "enum": ["get_question", "send_to_user", "summary_and_save", ""],
                      "description": "下一个状态。"
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
                              "const": "question"
                            },
                            "question_id": {
                              "type": "integer"
                            },
                            "question": {
                              "type": "string"
                            },
                            "remaining_questions": {
                              "type": "integer"
                            },
                            "current": {
                              "type": "integer"
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
                              "const": "evaluation"
                            },
                            "question_id": {
                              "type": "integer"
                            },
                            "scores": {
                              "type": "object",
                              "properties": {
                                "content_score": {
                                  "type": "integer",
                                  "minimum": 0,
                                  "maximum": 100
                                },
                                "coverage_score": {
                                  "type": "integer",
                                  "minimum": 0,
                                  "maximum": 100
                                },
                                "structure_score": {
                                  "type": "integer",
                                  "minimum": 0,
                                  "maximum": 100
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
                                  "items": {"type": "string"}
                                },
                                "missed_points": {
                                  "type": "array",
                                  "items": {"type": "string"}
                                },
                                "suggestion": {
                                  "type": "string"
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
                              "const": "summary"
                            },
                            "total_questions": {
                              "type": "integer"
                            },
                            "answered_questions": {
                              "type": "integer"
                            },
                            "overall_score": {
                              "type": "integer",
                              "minimum": 0,
                              "maximum": 100
                            },
                            "strengths": {
                              "type": "array",
                              "items": {"type": "string"}
                            },
                            "weaknesses": {
                              "type": "array",
                              "items": {"type": "string"}
                            },
                            "priority_actions": {
                              "type": "array",
                              "items": {"type": "string"}
                            }
                          },
                          "required": ["type", "total_questions", "answered_questions", "overall_score", "strengths", "weaknesses", "priority_actions"],
                          "additionalProperties": false
                        }
                      ]
                    }
                  },
                  "required": ["varName", "result", "nextState"],
                  "additionalProperties": false
                },
                {
                  "type": "object",
                  "description": "save_doc 的参数结构",
                  "properties": {
                    "varName": {
                      "type": "string",
                      "description": "变量名，固定为 InterviewHistory"
                    },
                    "type": {
                      "type": "string",
                      "enum": ["json"],
                      "description": "文档类型，固定为 json"
                    },
                    "content": {
                      "type": "string",
                      "description": "JSON数组字符串，包含所有历史记录"
                    },
                    "nextState": {
                      "type": "string",
                      "enum": ["get_question", "send_to_user", "summary_and_save", ""],
                      "description": "下一个状态。"
                    }
                  },
                  "required": ["varName", "type", "content", "nextState"],
                  "additionalProperties": false
                }
              ]
            }
          },
          "required": ["name", "arguments"]
        },
        "minItems": 1
      },
      "nextState": {
        "type": ["string", "null"],
        "enum": ["get_question", "send_to_user", "summary_and_save", "", null],
        "description": "最终的下一个状态。如果最后一个函数调用已经设置了 nextState，这里可以设置为 null 或空字符串。系统会使用最后一个函数调用的 nextState。"
      }
    },
    "required": ["content", "calls", "nextState"]
  }
}`
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
	elog.DefaultLogger.SetLevel(elog.DebugLevel)

	// 连接到 gRPC 服务器
	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("连接 gRPC 失败: %v", err)
	}
	defer conn.Close()

	client := chatv1.NewServiceClient(conn)
	mux := http.NewServeMux()

	// 注册面试相关路由
	registerInterviewRoutes(mux, client)

	// 注册音频转文本代理路由
	registerAudioProxyRoutes(mux, t)

	log.Println("HTTP 代理服务器启动于 :8080")
	log.Println("转发目标: localhost:9090 (gRPC)")
	log.Println("端点:")
	log.Println("   - POST /api/interview/chat/create")
	log.Println("   - POST /api/interview/stream")
	log.Println("   - POST /api/audio/transcriptions")
	log.Println("   - GET  /health")
	log.Println("按 Ctrl+C 停止服务器")
	log.Println("---")

	// 使用 http.Server 并设置timeout
	// WriteTimeout 设置为 0 表示不限制写入超时（SSE 流式响应需要长时间保持连接）
	// 或者设置为足够大的值（如 5 分钟）以支持 LLM 的长时间响应
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // 0 表示不限制，适合 SSE 长连接
		IdleTimeout:  60 * time.Second,
	}

	// 优雅关闭：监听信号并执行清理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 在 goroutine 中启动服务器
	errChan := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// 等待终止信号或服务器错误
	select {
	case err := <-errChan:
		t.Fatalf("HTTP 服务器启动失败: %v", err)
	case <-sigChan:
		log.Println("\n收到终止信号，正在优雅关闭 HTTP 服务器...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP 服务器关闭失败: %v", err)
		} else {
			log.Println("HTTP 服务器已优雅关闭")
		}
		// 测试函数返回，t.Cleanup 会自动执行
	}
}

// registerInterviewRoutes 注册面试相关的 HTTP 路由
func registerInterviewRoutes(mux *http.ServeMux, client chatv1.ServiceClient) {
	corsHandler := newCORSHandler()

	// 创建 Chat 接口
	mux.HandleFunc("/api/interview/chat/create", corsHandler(handleCreateChat(client)))

	// Stream 流式接口（SSE）
	mux.HandleFunc("/api/interview/stream", corsHandler(handleStream(client)))

	// 健康检查
	mux.HandleFunc("/health", handleHealth())
}

// newCORSHandler 创建 CORS 处理函数
func newCORSHandler() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
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
}

// handleCreateChat 处理创建 Chat 的请求
func handleCreateChat(client chatv1.ServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Uid   int64  `json:"uid"`
			Title string `json:"title"`
			BizId int64  `json:"biz_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("创建 Chat: uid=%d, title=%s, biz_id=%d", req.Uid, req.Title, req.BizId)

		// 调用 gRPC Save，传入 biz_id
		resp, err := client.Save(context.Background(), &chatv1.SaveRequest{
			Chat: &chatv1.Chat{
				Uid:   req.Uid,
				Title: req.Title,
			},
			BizId: req.BizId,
		})
		if err != nil {
			log.Printf("创建 Chat 失败: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Chat 已创建: %s", resp.Sn)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"chat_sn": resp.Sn,
		}); err != nil {
			log.Printf("编码响应失败: %v", err)
		}
	}
}

// handleStream 处理 Stream 流式请求
func handleStream(client chatv1.ServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ChatSn string `json:"chat_sn"`
			Input  string `json:"input"`
			Uid    int64  `json:"uid"`
			State  string `json:"state"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("收到请求: chat_sn=%s, input=%s (前30字), state=%s", req.ChatSn, truncate(req.Input, 30), req.State)

		// 调用 gRPC Stream，使用请求的上下文以便正确处理取消和超时
		streamRes, err := client.Stream(r.Context(), &chatv1.StreamRequest{
			ChatSn: req.ChatSn,
			Input: &chatv1.UserInput{
				Content: req.Input,
			},
			Uid:   req.Uid,
			Key:   "", // 未使用，传空字符串
			State: req.State,
		})
		if err != nil {
			log.Printf("调用 Stream 失败: %v", err)
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
			resp, err := streamRes.Recv()
			if err == io.EOF {
				fmt.Fprintf(w, "event: done\ndata: {}\n\n")
				flusher.Flush()
				log.Printf("Stream 完成 (共 %d 个 Delta 事件)", deltaCount)
				break
			}
			if err != nil {
				log.Printf("Stream 错误: %v", err)
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
				log.Printf("Delta #%d: %s", deltaCount, resp.GetDelta().Content)
			}
		}
	}
}

// handleHealth 处理健康检查请求
func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "未设置"
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"grpc":     "localhost:9090",
			"base_url": baseURL,
			"time":     time.Now().Format(time.RFC3339),
		})
	}
}

// registerAudioProxyRoutes 注册音频转文本代理路由
func registerAudioProxyRoutes(mux *http.ServeMux, t *testing.T) {
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
		t.Fatal("未设置 BASE_URL 环境变量")
	}

	origin := os.Getenv("ORIGIN")
	if origin == "" {
		t.Fatal("未设置 ORIGIN 环境变量")
	}

	corsHandler := newCORSHandler()
	proxyHandler := newAudioProxyHandler(apiKey, internalToken, baseURL, origin)

	// Audio API
	mux.HandleFunc("/api/audio/transcriptions", corsHandler(proxyHandler))
}

// newAudioProxyHandler 创建音频转文本代理处理器
func newAudioProxyHandler(apiKey, internalToken, baseURL, origin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := readRequestBody(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取请求体失败: %v", err), http.StatusBadRequest)
			return
		}

		targetURL := buildTargetURL(baseURL, r)
		logProxyRequest(r, bodyBytes, targetURL)

		req, err := createProxyRequest(r, targetURL, bodyBytes, apiKey, internalToken, origin)
		if err != nil {
			http.Error(w, fmt.Sprintf("创建请求失败: %v", err), http.StatusInternalServerError)
			return
		}

		resp, err := sendProxyRequest(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("请求失败: %v", err), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取响应失败: %v", err), http.StatusInternalServerError)
			return
		}

		logProxyResponse(resp, respBytes)
		copyResponseHeaders(w, resp)
		writeResponse(w, resp, respBytes)
	}
}

// readRequestBody 读取请求体
func readRequestBody(r *http.Request) ([]byte, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return bodyBytes, nil
}

// buildTargetURL 构建目标 URL
func buildTargetURL(baseURL string, r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/api")
	targetURL := baseURL + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}
	return targetURL
}

// logProxyRequest 记录代理请求日志
func logProxyRequest(r *http.Request, bodyBytes []byte, targetURL string) {
	log.Printf("代理请求: %s %s -> %s", r.Method, r.URL.Path, targetURL)
	if len(bodyBytes) > 0 && len(bodyBytes) < 2000 {
		log.Printf("请求体: %s", string(bodyBytes))
	}
}

// createProxyRequest 创建代理请求
// 使用原始请求的 context，它会随请求生命周期自动管理
func createProxyRequest(r *http.Request, targetURL string, bodyBytes []byte, apiKey, internalToken, origin string) (*http.Request, error) {
	// 直接使用原始请求的 context，它会随请求生命周期自动管理
	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	// 复制请求头
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-Internal-Token", internalToken)
	req.Header.Set("Origin", origin)
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	req.Header.Set("Content-Type", contentType)

	return req, nil
}

// sendProxyRequest 发送代理请求
// 使用请求的 context，它会自动处理超时和取消
func sendProxyRequest(req *http.Request) (*http.Response, error) {
	// 使用默认的 http.Client，它会使用请求中的 context
	// 不需要额外设置超时，因为 context 已经管理了生命周期
	client := &http.Client{}
	return client.Do(req)
}

// logProxyResponse 记录代理响应日志
func logProxyResponse(resp *http.Response, respBytes []byte) {
	log.Printf("响应状态: %d, 大小: %d 字节", resp.StatusCode, len(respBytes))
	if len(respBytes) < 2000 {
		log.Printf("响应体: %s", string(respBytes))
	}
}

// copyResponseHeaders 复制响应头
func copyResponseHeaders(w http.ResponseWriter, resp *http.Response) {
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
}

// writeResponse 写入响应
func writeResponse(w http.ResponseWriter, resp *http.Response, respBytes []byte) {
	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(respBytes); err != nil {
		log.Printf("写入响应失败: %v", err)
	}
}

// truncate 截断字符串用于日志显示
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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

// prepareKbaseTestData 准备Kbase测试数据
// 使用 go-elasticsearch 直接操作ES，创建索引并插入测试题目
func prepareKbaseTestData(t *testing.T) {
	// 1. 创建ES客户端
	esAddr := "http://localhost:9229"
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
			"question_id": 6,
			"level":       "junior",
			"title":       "请解释MySQL的主键和外键",
			"analysis":    "主键和外键是关系数据库的核心概念，用于唯一标识和建立表之间的关联。",
			"tags":        []string{"主键", "外键", "约束"},
			"answers": map[string]any{
				"15k": map[string]any{
					"content":    "主键(Primary Key)唯一标识表中的每一条记录，不能为NULL。外键(Foreign Key)用于建立表之间的关联关系，外键的值必须是另一个表的主键值或NULL。",
					"key_points": []string{"主键唯一性", "外键约束", "表关联"},
				},
				"25k": map[string]any{
					"content":    "主键选择策略：自增ID(简单高效)、UUID(全局唯一但无序)、业务主键(有业务含义)。外键的级联操作：ON DELETE CASCADE(级联删除)、ON UPDATE CASCADE(级联更新)。",
					"key_points": []string{"主键选择", "级联删除", "级联更新"},
				},
				"35k": map[string]any{
					"content":    "分布式场景下的主键设计：雪花算法(Snowflake)生成分布式ID，兼顾唯一性和趋势递增。外键的性能影响：外键会降低写入性能，高并发场景通常在应用层维护关联关系而不使用外键约束。",
					"key_points": []string{"分布式ID", "雪花算法", "外键性能", "应用层约束"},
				},
			},
			"created_at": now,
			"updated_at": now,
		},
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
