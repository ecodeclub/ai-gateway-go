package test

import (
	"context"
	_ "embed"
	"encoding/json"
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
	//go:embed testdata/resume_extraction_assistant.md
	resumeExtractionAssistantSystemPrompt string
)

func TestLLMHandler(t *testing.T) {
	baseURL := os.Getenv("BASE_URL")
	apiKey := os.Getenv("API_KEY")

	if apiKey == "" || baseURL == "" {
		t.Skip("set BASE_URL or API_KEY")
	}
	suite.Run(t, &TestLLMHandlerSuite{
		apiKey:  apiKey,
		baseURL: baseURL,
	})
}

type TestLLMHandlerSuite struct {
	suite.Suite
	apiKey  string
	baseURL string
	*testioc.TestApp
	handler      *openaihdl.Handler
	client       openai.Client
	configRepo   *repository.InvocationConfigRepo
	providerRepo *repository.ProviderRepository
}

func (s *TestLLMHandlerSuite) SetupSuite() {
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

func (s *TestLLMHandlerSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.DB.WithContext(ctx).Exec("TRUNCATE TABLE chats").Error
	s.NoError(err)
	err = s.DB.WithContext(ctx).Exec("TRUNCATE TABLE messages").Error
	s.NoError(err)
	s.TestApp.Rdb.FlushDB(ctx)
}

func (s *TestLLMHandlerSuite) TestClient() {
	t := s.T()
	t.Skip()

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

func (s *TestLLMHandlerSuite) newConfigVersion(t *testing.T) domain.InvocationConfigVersion {
	t.Helper()
	config := domain.InvocationConfig{
		Name:        "",
		Biz:         domain.BizConfig{ID: 1},
		Description: "",
	}
	id, err := s.configRepo.Save(t.Context(), config)
	require.NoError(t, err)

	config.ID = id
	version := domain.InvocationConfigVersion{
		Config:       config,
		Model:        domain.Model{ID: 1},
		Version:      "v0.1",
		Prompt:       "",
		SystemPrompt: "",
		JSONSchema:   resumeJSONSchema,
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
	vid, err := s.configRepo.SaveVersion(t.Context(), version)
	require.NoError(t, err)

	model, err := s.providerRepo.GetModel(t.Context(), 1)
	require.NoError(t, err)

	version.ID = vid
	version.Model = model
	return version
}

func (s *TestLLMHandlerSuite) TestUnmarshalFunc() {
	t := s.T()
	t.Skip()
	var p responses.FunctionToolParam
	for _, str := range []string{askUserJSON, emitJSON, invokeLLMJSON} {
		// t.Logf("json_str: %s\n", str)
		err := json.Unmarshal([]byte(str), &p)
		require.NoError(t, err)
		t.Logf("FunctionToolParam: %#v\n", p)
	}
}

func (s *TestLLMHandlerSuite) TestResponseID() {
	t := s.T()
	t.Skip()
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
	/*
	   llm_handler_test.go:285: respID: resp_68b92ad5fac881909aadba56f9b420a8044d7d2629c1cedb , text: 当然可以！有一天，两个番茄在路上散步。一个番茄不小心摔倒了，另一个番茄回过头说：“嘿，快点，别变成番茄酱了！”
	   llm_handler_test.go:335: respID: resp_68b92ad9a53c81908bd0008729da0af8044d7d2629c1cedb , text: 这个笑话利用了“双关”来制造幽默。一只番茄摔倒在地上，另一只番茄担心它会被压烂，变成番茄酱。这里的幽默来源于将摔倒后的番茄形象具体化为番茄酱——通常我们认为番茄酱需要经过特定的加工过程，而不是因为简单的摔倒，因此形成了一种荒谬感，让人觉得好笑。
	*/
}

func (s *TestLLMHandlerSuite) TestHandler_Stream() {
	t := s.T()
	// t.Skip()

	_, err := s.providerRepo.SaveProvider(t.Context(), domain.Provider{
		ID:     1,
		Name:   "openai",
		APIKey: "fake-key",
	})
	require.NoError(t, err)

	_, err = s.providerRepo.SaveModel(t.Context(), domain.Model{
		ID:          1,
		Provider:    domain.Provider{ID: 1},
		Name:        "gpt-4o",
		InputPrice:  10,
		OutputPrice: 100,
		PriceMode:   "fake-price-mode",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err = s.DB.Model(&dao.Provider{}).Where("id = ?", 1).Error
		require.NoError(t, err)
		err = s.DB.Model(&dao.Model{}).Where("id = ?", 1).Error
		require.NoError(t, err)
	})

	testCases := []struct {
		name         string
		reqFunc      func(t *testing.T) domain.StreamRequest
		handleEvents func(t *testing.T, req domain.StreamRequest, events chan domain.StreamEvent)
	}{
		{
			name: "ask_user_讲笑话",
			reqFunc: func(t *testing.T) domain.StreamRequest {
				t.Helper()
				return domain.StreamRequest{
					ConfigVersion: s.newConfigVersion(t),
					Messages: []domain.Message{
						{
							Role:    domain.SYSTEM,
							Content: "你是一个笑话大王，可以针对不同性别、年龄的用户讲针对性的笑话。询问用户的年龄和性别等可以使用ask_user函数调用。你必须在一次函数调用中问清楚所有问题。",
						},
						{
							Role:    domain.USER,
							Content: "请给我讲一个笑话。",
						},
					},
				}
			},
			handleEvents: func(t *testing.T, req domain.StreamRequest, events chan domain.StreamEvent) {
				t.Helper()
				previousResponseID, callID := s.logEvents(t, events)
				events, err := s.handler.Stream(t.Context(), domain.StreamRequest{
					CallID:             callID,
					PreviousResponseID: previousResponseID,
					ConfigVersion:      req.ConfigVersion,
					Messages: []domain.Message{
						{
							Role:    domain.USER,
							Content: "男，25岁",
						},
					},
				})
				require.NoError(t, err)
				s.logEvents(t, events)
			},
		},

		{
			name: "ask_user_讲笑话2",
			reqFunc: func(t *testing.T) domain.StreamRequest {
				t.Helper()
				config := domain.InvocationConfig{
					Name:        "讲笑话",
					Biz:         domain.BizConfig{ID: 1},
					Description: "",
				}
				id, err := s.configRepo.Save(t.Context(), config)
				require.NoError(t, err)
				config.ID = id
				version := domain.InvocationConfigVersion{
					Config:       config,
					Model:        domain.Model{ID: 1},
					Version:      "v0.1",
					Prompt:       "",
					SystemPrompt: "你是一个笑话大王，可以针对不同性别、年龄的用户讲针对性的笑话。询问用户的年龄和性别等可以使用ask_user函数调用。你必须在一次函数调用中问清楚所有问题。",
					Functions: []domain.Function{
						{
							Name:       fcall.NameAskUser,
							Definition: askUserJSON,
						},
					},
					Temperature: 0,
					TopP:        0,
					MaxTokens:   2500,
					Status:      domain.InvocationCfgVersionStatusActive,
				}
				vid, err := s.configRepo.SaveVersion(t.Context(), version)
				require.NoError(t, err)
				model, err := s.providerRepo.GetModel(t.Context(), 1)
				require.NoError(t, err)
				version.ID = vid
				version.Model = model
				return domain.StreamRequest{
					ConfigVersion: version,
					Messages: []domain.Message{
						{
							Role:    domain.USER,
							Content: "请给我讲一个笑话。",
						},
						{
							Role:    domain.SYSTEM,
							Content: "讲完笑话后，你还有说明一下这个笑话为什么好笑。",
						},
					},
				}
			},
			handleEvents: func(t *testing.T, req domain.StreamRequest, events chan domain.StreamEvent) {
				t.Helper()
				previousResponseID, callID := s.logEvents(t, events)
				events, err := s.handler.Stream(t.Context(), domain.StreamRequest{
					CallID:             callID,
					PreviousResponseID: previousResponseID,
					ConfigVersion:      req.ConfigVersion,
					Messages: []domain.Message{
						{
							Role:    domain.USER,
							Content: "男，25岁",
						},
					},
				})
				require.NoError(t, err)
				s.logEvents(t, events)
			},
		},

		// {
		// 	name: "ask_user_简历信息提取",
		// 	reqFunc: func(t *testing.T) domain.StreamRequest {
		// 		t.Helper()
		// 		config := domain.InvocationConfig{
		// 			Name:        "简历信息提取",
		// 			Biz:         domain.BizConfig{ID: 1},
		// 			Description: "",
		// 		}
		// 		id, err := s.configRepo.Save(t.Context(), config)
		// 		require.NoError(t, err)
		// 		config.ID = id
		// 		version := domain.InvocationConfigVersion{
		// 			Config:       config,
		// 			Model:        domain.Model{ID: 1},
		// 			Version:      "v0.1",
		// 			Prompt:       "",
		// 			SystemPrompt: resumeExtractionAssistantSystemPrompt,
		// 			JSONSchema:   resumeJSONSchema,
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
		// 		vid, err := s.configRepo.SaveVersion(t.Context(), version)
		// 		require.NoError(t, err)
		// 		model, err := s.providerRepo.GetModel(t.Context(), 1)
		// 		require.NoError(t, err)
		// 		version.ID = vid
		// 		version.Model = model
		// 		return domain.StreamRequest{
		// 			ConfigVersion: version,
		// 			Messages: []domain.Message{
		// 				{
		// 					Role:    domain.USER,
		// 					Content: resumeXiaoMing,
		// 				},
		// 			},
		// 		}
		// 	},
		// 	handleEvents: func(t *testing.T, req domain.StreamRequest, events chan domain.StreamEvent) {
		// 		t.Helper()
		// 		s.logEvents(t, events)
		// 	},
		// },
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.reqFunc(t)
			version := req.ConfigVersion
			t.Cleanup(func() {
				err = s.DB.Model(&dao.InvocationConfig{}).Where("id = ?", version.Config.ID).Error
				require.NoError(t, err)
				err = s.DB.Model(&dao.InvocationConfigVersion{}).Where("id = ?", version.ID).Error
				require.NoError(t, err)
			})
			events, err1 := s.handler.Stream(t.Context(), req)
			require.NoError(t, err1)
			tc.handleEvents(t, req, events)
		})
	}

}

func (s *TestLLMHandlerSuite) logEvents(t *testing.T, events chan domain.StreamEvent) (respID, callID string) {
	defer func() {
		log.Printf("退出本次logEvents\n")
	}()
	for event := range events {
		if event.Done {
			return
		}
		if event.CallID != "" {
			callID = event.CallID
		}
		if event.ResponseID != "" {
			respID = event.ResponseID
		}
		t.Logf("event = %#v\n", event)
		// t.Logf("ConversationID: %s\n", event.ConversationID)
		// t.Logf("ResponseID: %s\n", event.ResponseID)
		// t.Logf("CallID: %s\n", event.CallID)
		// t.Logf("ReasoningContent: %s\n", event.ReasoningContent)
		// t.Logf("Content: %s\n", event.Content)
		// t.Logf("InputToken: %d\n", event.InputToken)
		// t.Logf("OutputToken: %d\n", event.OutputToken)
		// t.Logf("Error: %s\n", event.Error)
	}
	return
}
