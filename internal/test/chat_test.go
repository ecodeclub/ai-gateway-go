package test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ecodeclub/ai-gateway-go/cmd/platform/ioc"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"
	openaihdl "github.com/ecodeclub/ai-gateway-go/internal/service/llm/platform/openai"
	testioc "github.com/ecodeclub/ai-gateway-go/internal/test/ioc"
	"github.com/ecodeclub/ekit/slice"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/conversations"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	//go:embed testdata/ask_user.json
	askUserJSON string
	//go:embed testdata/emit_json.json
	emitJSON string
	//go:embed testdata/invoke_llm.json
	invokeLLMJSON string
	//go:embed testdata/resume_json_schema.json
	resumeJSONSchema string
	//go:embed testdata/resume_xiaoming.md
	resumeXiaoMing string
	//go:embed testdata/resume_xiaoming_missing_info.md
	resumeXiaoMingMissingInfo string
	//go:embed testdata/resume_extraction_assistant.md
	resumeExtractionAssistantSystemPrompt string
	//go:embed testdata/resume_extraction_assistant_v2.md
	resumeExtractionAssistantSystemPromptV2 string
	//go:embed testdata/resume_evaluation_assistant.md
	resumeEvaluationAssistantSystemPrompt string
)

func TestChatTestSuite(t *testing.T) {
	baseURL := os.Getenv("BASE_URL")
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" || baseURL == "" {
		t.Skip("BASE_URL 和/或 API_KEY 环境变量未设置")
	}
	suite.Run(t, &ChatTestSuite{
		apiKey:  apiKey,
		baseURL: baseURL,
	})
}

type ChatTestSuite struct {
	suite.Suite
	apiKey  string
	baseURL string
	*testioc.TestApp
	handler      *openaihdl.Handler
	client       openai.Client
	configRepo   *repository.InvocationConfigRepo
	providerRepo *repository.ProviderRepository
}

func (s *ChatTestSuite) SetupSuite() {
	db := testioc.InitDB()
	s.configRepo = repository.NewInvocationConfigRepo(dao.NewInvocationConfigDAO(db))
	s.providerRepo = repository.NewProviderRepository(dao.NewProviderDAO(db))
	invokeLLMCall := fcall.NewInvokeLLMFuncCall(
		s.configRepo,
		ioc.InitRender(),
	)
	registry := ioc.InitFunctionCallRegistry(
		fcall.NewAskUserFunctionCall(),
		fcall.NewEmitJsonFunctionCall(),
		invokeLLMCall,
	)

	s.client = openai.NewClient(
		option.WithBaseURL(s.baseURL),
		option.WithAPIKey(s.apiKey),
	)
	s.handler = openaihdl.NewHandler(s.client, registry)
	s.TestApp = testioc.InitApp(testioc.TestOnly{LLM: s.handler})
}

func (s *ChatTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.DB.WithContext(ctx).Exec("TRUNCATE TABLE chats").Error
	s.NoError(err)
	err = s.DB.WithContext(ctx).Exec("TRUNCATE TABLE messages").Error
	s.NoError(err)
	err = s.DB.WithContext(ctx).Exec("TRUNCATE TABLE temp_quotas").Error
	s.NoError(err)
	err = s.DB.WithContext(ctx).Exec("TRUNCATE TABLE quotas").Error
	s.NoError(err)
	err = s.DB.WithContext(ctx).Exec("TRUNCATE TABLE quota_records").Error
	s.NoError(err)
	s.TestApp.Rdb.FlushDB(ctx)
}

func (s *ChatTestSuite) TestClient() {
	t := s.T()
	t.Skip("测试 BASE_URL 和 API_KEY 环境变量设置是否正确")

	params := responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: slice.Map([]domain.Message{
				{
					Role:    domain.USER,
					Content: "给我讲个笑话",
				},
			}, func(idx int, src domain.Message) responses.ResponseInputItemUnionParam {
				contentList := responses.ResponseInputMessageContentListParam{
					{
						OfInputText: &responses.ResponseInputTextParam{Text: src.Content},
					},
				}
				switch src.Role {
				case domain.USER:
					return responses.ResponseInputItemUnionParam{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Content: contentList,
							Role:    "user",
						},
					}
				case domain.SYSTEM:
					return responses.ResponseInputItemUnionParam{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Content: contentList,
							Role:    "system",
						},
					}
				default:
					return responses.ResponseInputItemUnionParam{}
				}
			}),
		},
		Model: "gpt-4o",
	}

	resp, err := s.client.Responses.New(t.Context(), params)
	require.NoError(t, err)

	t.Logf("输出：%s", resp.OutputText())

	params.PreviousResponseID = openai.String(resp.ID)
	params.Input = responses.ResponseNewParamsInputUnion{}

	params.Input.OfInputItemList = append(params.Input.OfInputItemList, responses.ResponseInputItemUnionParam{
		OfInputMessage: &responses.ResponseInputItemMessageParam{
			Role: "user",
			Content: []responses.ResponseInputContentUnionParam{{
				OfInputText: &responses.ResponseInputTextParam{
					Text: "解释一下这个笑话为什么好笑？",
				},
			}},
		},
	})

	resp, err = s.client.Responses.New(t.Context(), params)
	require.NoError(t, err)

	t.Logf("输出：%s", resp.OutputText())
}

func (s *ChatTestSuite) TestUnmarshalFunctionToolParam() {
	t := s.T()
	t.Skip("测试定义的Function Definition是否能够被正确反序列化")
	var p responses.FunctionToolParam
	for _, str := range []string{askUserJSON, emitJSON, invokeLLMJSON} {
		// t.Logf("json_str: %s\n", str)
		err := json.Unmarshal([]byte(str), &p)
		require.NoError(t, err)
		t.Logf("FunctionToolParam: %#v\n", p)
	}
}

func (s *ChatTestSuite) TestPreviousResponseID() {
	t := s.T()
	t.Skip("测试Responses API的PreviousResponseID特性，有效但受模型影响较大，且不太稳定")
	params := responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: slice.Map([]domain.Message{
				{
					Role:    domain.USER,
					Content: "给我讲个笑话",
				},
			}, func(idx int, src domain.Message) responses.ResponseInputItemUnionParam {
				contentList := responses.ResponseInputMessageContentListParam{
					{
						OfInputText: &responses.ResponseInputTextParam{Text: src.Content},
					},
				}
				switch src.Role {
				case domain.USER:
					return responses.ResponseInputItemUnionParam{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Content: contentList,
							Role:    "user",
						},
					}
				case domain.SYSTEM:
					return responses.ResponseInputItemUnionParam{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Content: contentList,
							Role:    "system",
						},
					}
				default:
					return responses.ResponseInputItemUnionParam{}
				}
			}),
		},
		Model: "gpt-4o",
	}
	stream := s.client.Responses.NewStreaming(t.Context(), params)
	var responseID string
	for stream.Next() {
		event := stream.Current()
		switch event.Type {
		case "error":
		case "response.created":
			responseID = event.Response.ID
		case "response.completed":
			assert.Equal(t, responseID, event.Response.ID)
		case "response.output_text.done":
			text := event.AsResponseOutputTextDone()
			t.Logf("respID: %s , text: %s", responseID, text.Text)
		}
	}
	params = responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: slice.Map([]domain.Message{
				{
					Role:    domain.USER,
					Content: "解释一下为什么好笑？",
				},
			}, func(idx int, src domain.Message) responses.ResponseInputItemUnionParam {
				contentList := responses.ResponseInputMessageContentListParam{
					{
						OfInputText: &responses.ResponseInputTextParam{Text: src.Content},
					},
				}
				switch src.Role {
				case domain.USER:
					return responses.ResponseInputItemUnionParam{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Content: contentList,
							Role:    "user",
						},
					}
				case domain.SYSTEM:
					return responses.ResponseInputItemUnionParam{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Content: contentList,
							Role:    "system",
						},
					}
				default:
					return responses.ResponseInputItemUnionParam{}
				}
			}),
		},
		Model:              "gpt-4o",
		PreviousResponseID: openai.String(responseID),
	}
	stream = s.client.Responses.NewStreaming(t.Context(), params)
	for stream.Next() {
		event := stream.Current()
		switch event.Type {
		case "error":
		case "response.created":
			responseID = event.Response.ID
		case "response.completed":
			assert.Equal(t, responseID, event.Response.ID)
		case "response.output_text.done":
			text := event.AsResponseOutputTextDone()
			t.Logf("respID: %s , text: %s", responseID, text.Text)
		}
	}
}

func (s *ChatTestSuite) TestConversationsAPI() {
	t := s.T()
	t.Skip("测试Conversations API的Conversation特性，更换了多个中转商都不支持该接口")
	cvs, err := s.client.Conversations.New(t.Context(), conversations.ConversationNewParams{})
	require.NoError(t, err)

	params := responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: slice.Map([]domain.Message{
				{
					Role:    domain.USER,
					Content: "给我讲个笑话",
				},
			}, func(idx int, src domain.Message) responses.ResponseInputItemUnionParam {
				contentList := responses.ResponseInputMessageContentListParam{
					{
						OfInputText: &responses.ResponseInputTextParam{Text: src.Content},
					},
				}
				switch src.Role {
				case domain.USER:
					return responses.ResponseInputItemUnionParam{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Content: contentList,
							Role:    "user",
						},
					}
				case domain.SYSTEM:
					return responses.ResponseInputItemUnionParam{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Content: contentList,
							Role:    "system",
						},
					}
				default:
					return responses.ResponseInputItemUnionParam{}
				}
			}),
		},
		Model: "gpt-4o-mini",
		Conversation: responses.ResponseNewParamsConversationUnion{
			OfConversationObject: &responses.ResponseConversationParam{
				ID: cvs.ID,
			},
		},
	}

	resp, err := s.client.Responses.New(t.Context(), params)
	require.NoError(t, err)

	t.Logf("输出：%s", resp.OutputText())

	params.PreviousResponseID = openai.String(resp.ID)
	params.Input = responses.ResponseNewParamsInputUnion{}
	params.Input.OfInputItemList = append(params.Input.OfInputItemList, responses.ResponseInputItemUnionParam{
		OfInputMessage: &responses.ResponseInputItemMessageParam{
			Role: "user",
			Content: []responses.ResponseInputContentUnionParam{{
				OfInputText: &responses.ResponseInputTextParam{
					Text: "解释一下这个笑话为什么好笑？",
				},
			}},
		},
	})

	resp, err = s.client.Responses.New(t.Context(), params)
	require.NoError(t, err)

	t.Logf("输出：%s", resp.OutputText())
}

func (s *ChatTestSuite) consumeEvents(t *testing.T, events chan domain.StreamEvent) {
	t.Helper()
	defer func() {
		log.Printf("退出本次logEvents\n")
	}()
	for event := range events {
		if event.Done {
			return
		}
		// t.Logf("event = %#v\n", event)
	}
}

func (s *ChatTestSuite) TestChatService_Stream() {
	t := s.T()

	uid := int64(1890521)
	userKey := fmt.Sprintf("key-%d", uid)
	err := s.QuotaService.AddQuota(t.Context(), domain.Quota{
		Amount: 10000,
		Key:    userKey,
		Uid:    uid,
	})
	require.NoError(t, err)

	providerID := int64(2)
	modelID := int64(2)
	_, err = s.providerRepo.SaveProvider(t.Context(), domain.Provider{
		ID:     providerID,
		Name:   "openai",
		APIKey: "fake-key",
	})
	require.NoError(t, err)

	_, err = s.providerRepo.SaveModel(t.Context(), domain.Model{
		ID:          modelID,
		Provider:    domain.Provider{ID: providerID},
		Name:        "gpt-4o",
		InputPrice:  10,
		OutputPrice: 100,
		PriceMode:   "fake-price-mode",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err = s.DB.Model(&dao.Provider{}).Where("id = ?", providerID).Error
		require.NoError(t, err)
		err = s.DB.Model(&dao.Model{}).Where("id = ?", modelID).Error
		require.NoError(t, err)
	})

	testCases := []struct {
		name         string
		reqFunc      func(t *testing.T) domain.ChatStreamRequest
		handleEvents func(t *testing.T, req domain.ChatStreamRequest, events chan domain.StreamEvent)
	}{
		// {
		// 	name: "ask_user_讲笑话",
		// 	reqFunc: func(t *testing.T) domain.ChatStreamRequest {
		// 		t.Helper()
		// 		sn, err1 := s.ChatService.Save(t.Context(), domain.Chat{
		// 			Uid:   uid,
		// 			Title: "讲笑话",
		// 		})
		// 		require.NoError(t, err1)
		//
		// 		config := domain.InvocationConfig{
		// 			Name:        "",
		// 			Biz:         domain.BizConfig{ID: 1},
		// 			Description: "",
		// 		}
		// 		id, err1 := s.configRepo.Save(t.Context(), config)
		// 		require.NoError(t, err1)
		//
		// 		config.ID = id
		// 		version := domain.InvocationConfigVersion{
		// 			Config:     config,
		// 			Model:      domain.Model{ID: modelID},
		// 			Version:    "v0.1",
		// 			Attributes: nil,
		// 			Functions: []domain.Function{
		// 				{
		// 					Name:       fcall.NameAskUser,
		// 					Definition: askUserJSON,
		// 				},
		// 			},
		// 			Temperature: 0,
		// 			TopP:        0,
		// 			MaxTokens:   2500,
		// 			Status:      domain.InvocationCfgVersionStatusActive,
		// 		}
		// 		vid, err1 := s.configRepo.SaveVersion(t.Context(), version)
		// 		require.NoError(t, err1)
		// 		model, err1 := s.providerRepo.GetModel(t.Context(), 1)
		// 		require.NoError(t, err1)
		// 		version.ID = vid
		// 		version.Model = model
		// 		return domain.ChatStreamRequest{
		// 			Sn: sn,
		// 			Messages: []domain.Message{
		// 				{
		// 					Role:    domain.SYSTEM,
		// 					Content: "你是一个笑话大王，可以针对不同性别、年龄的用户讲针对性的笑话。询问用户的年龄和性别等可以使用ask_user函数调用。你必须在一次函数调用中问清楚所有问题。",
		// 				},
		// 				{
		// 					Role:    domain.USER,
		// 					Content: "请给我讲一个笑话。",
		// 				},
		// 			},
		// 			InvocationConfigID: config.ID,
		// 			Uid:                uid,
		// 			Key:                userKey,
		// 		}
		// 	},
		// 	handleEvents: func(t *testing.T, req domain.ChatStreamRequest, events chan domain.StreamEvent) {
		// 		t.Helper()
		// 		s.consumeEvents(t, events)
		// 		events, err := s.ChatService.Stream(t.Context(), domain.ChatStreamRequest{
		// 			Sn:                 req.Sn,
		// 			InvocationConfigID: req.InvocationConfigID,
		// 			Messages: []domain.Message{
		// 				{
		// 					Role:    domain.USER,
		// 					Content: "男，25岁",
		// 				},
		// 			},
		// 		})
		// 		require.NoError(t, err)
		// 		s.consumeEvents(t, events)
		//
		// 		chat, err := s.ChatService.Detail(t.Context(), req.Sn)
		// 		require.NoError(t, err)
		// 		for i := range chat.Messages {
		// 			t.Logf("chat[%s].Message[%d]: %#v\n", req.Sn, i, chat.Messages[i])
		// 		}
		// 	},
		// },
		// {
		// 	name: "ask_user_使用systemPrompt讲笑话",
		// 	reqFunc: func(t *testing.T) domain.ChatStreamRequest {
		// 		t.Helper()
		// 		sn, err1 := s.ChatService.Save(t.Context(), domain.Chat{
		// 			Uid:   uid,
		// 			Title: "使用systemPrompt讲笑话",
		// 		})
		// 		require.NoError(t, err1)
		//
		// 		config := domain.InvocationConfig{
		// 			Name:        "使用systemPrompt讲笑话",
		// 			Biz:         domain.BizConfig{ID: 1},
		// 			Description: "",
		// 		}
		// 		id, err1 := s.configRepo.Save(t.Context(), config)
		// 		require.NoError(t, err1)
		//
		// 		config.ID = id
		// 		version := domain.InvocationConfigVersion{
		// 			Config:  config,
		// 			Model:   domain.Model{ID: modelID},
		// 			Version: "v0.1",
		// 			// 用户写的
		// 			Prompt: "",
		// 			// 我们写的Prompt
		// 			SystemPrompt: "你是一个笑话大王，可以针对不同性别、年龄的用户讲针对性的笑话。询问用户的年龄和性别等可以使用ask_user函数调用。你必须在一次函数调用中问清楚所有问题。",
		// 			Attributes:   nil,
		// 			Functions: []domain.Function{
		// 				{
		// 					Name:       fcall.NameAskUser,
		// 					Definition: askUserJSON,
		// 				},
		// 			},
		// 			Temperature: 0,
		// 			TopP:        0,
		// 			MaxTokens:   2500,
		// 			Status:      domain.InvocationCfgVersionStatusActive,
		// 		}
		// 		vid, err1 := s.configRepo.SaveVersion(t.Context(), version)
		// 		require.NoError(t, err1)
		// 		model, err1 := s.providerRepo.GetModel(t.Context(), 1)
		// 		require.NoError(t, err1)
		// 		version.ID = vid
		// 		version.Model = model
		// 		return domain.ChatStreamRequest{
		// 			Sn: sn,
		// 			Messages: []domain.Message{
		// 				{
		// 					Role:    domain.USER,
		// 					Content: "请给我讲一个笑话。",
		// 				},
		// 			},
		// 			InvocationConfigID: config.ID,
		// 			Uid:                uid,
		// 			Key:                userKey,
		// 		}
		// 	},
		// 	handleEvents: func(t *testing.T, req domain.ChatStreamRequest, events chan domain.StreamEvent) {
		// 		t.Helper()
		// 		s.consumeEvents(t, events)
		// 		events, err := s.ChatService.Stream(t.Context(), domain.ChatStreamRequest{
		// 			Sn:                 req.Sn,
		// 			InvocationConfigID: req.InvocationConfigID,
		// 			Messages: []domain.Message{
		// 				{
		// 					Role:    domain.USER,
		// 					Content: "男，25岁",
		// 				},
		// 			},
		// 		})
		// 		require.NoError(t, err)
		// 		s.consumeEvents(t, events)
		//
		// 		chat, err := s.ChatService.Detail(t.Context(), req.Sn)
		// 		require.NoError(t, err)
		// 		for i := range chat.Messages {
		// 			t.Logf("chat[%s].Message[%d]: %#v\n", req.Sn, i, chat.Messages[i])
		// 		}
		// 	},
		// },
		// {
		// 	name: "ask_user_emit_json_提取简历信息",
		// 	reqFunc: func(t *testing.T) domain.ChatStreamRequest {
		// 		t.Helper()
		// 		sn, err1 := s.ChatService.Save(t.Context(), domain.Chat{
		// 			Uid:   uid,
		// 			Title: "提取简历信息",
		// 		})
		// 		require.NoError(t, err1)
		//
		// 		config := domain.InvocationConfig{
		// 			Name:        "提取简历信息",
		// 			Biz:         domain.BizConfig{ID: 1},
		// 			Description: "",
		// 		}
		// 		id, err1 := s.configRepo.Save(t.Context(), config)
		// 		require.NoError(t, err1)
		//
		// 		config.ID = id
		// 		version := domain.InvocationConfigVersion{
		// 			Config:       config,
		// 			Model:        domain.Model{ID: modelID},
		// 			Version:      "v0.1",
		// 			Prompt:       "",
		// 			SystemPrompt: resumeExtractionAssistantSystemPrompt,
		// 			Attributes:   nil,
		// 			Functions: []domain.Function{
		// 				{
		// 					Name:       fcall.NameAskUser,
		// 					Definition: askUserJSON,
		// 				},
		// 				{
		// 					Name:       fcall.NameEmitJSON,
		// 					Definition: emitJSON,
		// 				},
		// 			},
		// 			Temperature: 0,
		// 			TopP:        0,
		// 			MaxTokens:   2500,
		// 			Status:      domain.InvocationCfgVersionStatusActive,
		// 		}
		// 		vid, err1 := s.configRepo.SaveVersion(t.Context(), version)
		// 		require.NoError(t, err1)
		// 		model, err1 := s.providerRepo.GetModel(t.Context(), 1)
		// 		require.NoError(t, err1)
		// 		version.ID = vid
		// 		version.Model = model
		// 		return domain.ChatStreamRequest{
		// 			Sn: sn,
		// 			Messages: []domain.Message{
		// 				{
		// 					Role:    domain.USER,
		// 					Content: resumeXiaoMingMissingInfo,
		// 				},
		// 			},
		// 			InvocationConfigID: config.ID,
		// 			Uid:                uid,
		// 			Key:                userKey,
		// 		}
		// 	},
		// 	handleEvents: func(t *testing.T, req domain.ChatStreamRequest, events chan domain.StreamEvent) {
		// 		t.Helper()
		// 		infos := []string{"男，25岁", "13823456789", "lixiaoming@example.com", "2年工作经验，期望月薪12-18K"}
		// 		for i := range infos {
		// 			s.consumeEvents(t, events)
		// 			events, err = s.ChatService.Stream(t.Context(), domain.ChatStreamRequest{
		// 				Sn:                 req.Sn,
		// 				InvocationConfigID: req.InvocationConfigID,
		// 				Messages: []domain.Message{
		// 					{
		// 						Role:    domain.USER,
		// 						Content: infos[i],
		// 					},
		// 				},
		// 			})
		// 			require.NoError(t, err)
		// 		}
		// 		s.consumeEvents(t, events)
		// 		chat, err := s.ChatService.Detail(t.Context(), req.Sn)
		// 		require.NoError(t, err)
		// 		for i := range chat.Messages {
		// 			t.Logf("chat[%s].Message[%d]: %#v\n", req.Sn, i, chat.Messages[i])
		// 		}
		// 	},
		// },
		{
			name: "ask_user_emit_json_invoke_llm_评价简历",
			reqFunc: func(t *testing.T) domain.ChatStreamRequest {
				t.Helper()
				sn, err1 := s.ChatService.Save(t.Context(), domain.Chat{
					Uid:   uid,
					Title: "提取简历信息",
				})
				require.NoError(t, err1)

				config := domain.InvocationConfig{
					Name:        "提取简历信息V2",
					Biz:         domain.BizConfig{ID: 1},
					Description: "",
				}
				id, err1 := s.configRepo.Save(t.Context(), config)
				require.NoError(t, err1)

				config.ID = id
				version := domain.InvocationConfigVersion{
					Config:       config,
					Model:        domain.Model{ID: modelID},
					Version:      "v0.2",
					Prompt:       "",
					SystemPrompt: resumeExtractionAssistantSystemPromptV2,
					Attributes:   nil,
					Functions: []domain.Function{
						{
							Name:       fcall.NameAskUser,
							Definition: askUserJSON,
						},
						{
							Name:       fcall.NameEmitJSON,
							Definition: emitJSON,
						},
						{
							Name:       fcall.NameInvokeLLM,
							Definition: invokeLLMJSON,
						},
					},
					Temperature: 0,
					TopP:        0,
					MaxTokens:   2500,
					Status:      domain.InvocationCfgVersionStatusActive,
				}
				vid, err1 := s.configRepo.SaveVersion(t.Context(), version)
				require.NoError(t, err1)
				model, err1 := s.providerRepo.GetModel(t.Context(), 1)
				require.NoError(t, err1)
				version.ID = vid
				version.Model = model

				config2 := domain.InvocationConfig{
					ID:          12345,
					Name:        "评价简历",
					Biz:         domain.BizConfig{ID: 1},
					Description: "",
				}
				id, err1 = s.configRepo.Save(t.Context(), config2)
				require.NoError(t, err1)

				version2 := domain.InvocationConfigVersion{
					Config:       config2,
					Model:        domain.Model{ID: modelID},
					Version:      "v0.1",
					Prompt:       "",
					SystemPrompt: resumeEvaluationAssistantSystemPrompt,
					Temperature:  0,
					TopP:         0,
					MaxTokens:    2500,
					Status:       domain.InvocationCfgVersionStatusActive,
				}
				vid, err1 = s.configRepo.SaveVersion(t.Context(), version2)
				require.NoError(t, err1)

				return domain.ChatStreamRequest{
					Sn: sn,
					Messages: []domain.Message{
						{
							Role:    domain.USER,
							Content: resumeXiaoMing,
						},
					},
					InvocationConfigID: config.ID,
					Uid:                uid,
					Key:                userKey,
				}
			},
			handleEvents: func(t *testing.T, req domain.ChatStreamRequest, events chan domain.StreamEvent) {
				t.Helper()
				s.consumeEvents(t, events)
				events, err = s.ChatService.Stream(t.Context(), domain.ChatStreamRequest{
					Sn:                 req.Sn,
					InvocationConfigID: req.InvocationConfigID,
					Messages: []domain.Message{
						{
							Role:    domain.USER,
							Content: "初级",
						},
					},
				})
				require.NoError(t, err)
				s.consumeEvents(t, events)
				chat, err := s.ChatService.Detail(t.Context(), req.Sn)
				require.NoError(t, err)
				for i := range chat.Messages {
					t.Logf("chat[%s].Message[%d]: %#v\n", req.Sn, i, chat.Messages[i])
				}
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.reqFunc(t)
			t.Cleanup(func() {
				err = s.DB.Model(&dao.InvocationConfig{}).Where("id = ?", req.InvocationConfigID).Error
				require.NoError(t, err)
				err = s.DB.Model(&dao.InvocationConfigVersion{}).Where("inv_id = ?", req.InvocationConfigID).Error
				require.NoError(t, err)
			})
			events, err1 := s.ChatService.Stream(t.Context(), req)
			require.NoError(t, err1)
			tc.handleEvents(t, req, events)
		})
	}
}
