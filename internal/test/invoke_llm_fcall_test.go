package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"
	"github.com/ecodeclub/ai-gateway-go/internal/service/mocks"
	testioc "github.com/ecodeclub/ai-gateway-go/internal/test/ioc"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

type InvokeLLmFcall struct {
	suite.Suite
	*testioc.TestApp
	handler *mocks.MockHandler
	db      *gorm.DB
}

func (i *InvokeLLmFcall) SetupSuite() {
	ctrl := gomock.NewController(i.T())
	handler := mocks.NewMockHandler(ctrl)
	i.handler = handler
	app := testioc.InitApp(testioc.TestOnly{LLM: handler})
	i.TestApp = app
	db := testioc.InitDB()
	i.db = db
}

func TestInvokeLLmFcall(t *testing.T) { suite.Run(t, new(InvokeLLmFcall)) }

// Call 在参数反序列化出错时应返回错误，并且不应写入附件
func (i *InvokeLLmFcall) TestCall() {
	ctx := i.T().Context()
	// 构造上下文与请求参数（即便是合法 JSON，根据当前实现会返回反序列化错误）
	jsonMap := map[string]any{
		"level": "advanced",
	}
	i.handler.EXPECT().Chat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, messages []domain.Message) (domain.ChatResponse, error) {
		assert.Equal(i.T(), []domain.Message{
			{
				Role: domain.USER,
				Content: `这是我的评判标准，你需要对用户的简历信息进行一个评判，而后输出
1. 如果简历信息达到要求，则执行下一个 llm_invoke
2. 否则，执行ask_user，要求用户补充更多
评判标准
需要有高并发高可用的工作经历
`,
			},
		}, messages)
		return domain.ChatResponse{
			Response: domain.Message{
				Role:    domain.USER,
				Content: "ans",
			},
		}, nil
	}).AnyTimes()
	i.initVersion()
	fctx := &fcall.Context{Context: ctx, JSONData: jsonMap, Attachments: map[string]string{}}
	req := fcall.Request{Args: []byte(`{"invocation_id":"1000","gjson":"engineer"}`)}
	_, err := i.InvokeLLmFcall.Call(fctx, req)
	require.NoError(i.T(), err)
	// 校验fctx中的数据
	resp := domain.ChatResponse{
		Response: domain.Message{
			Role:    domain.USER,
			Content: "ans",
		},
	}
	respByte, err := json.Marshal(resp)
	require.NoError(i.T(), err)
	assert.Equal(i.T(), fctx.Attachments, map[string]string{
		"invoke_llm": string(respByte),
	})
}

func (i *InvokeLLmFcall) initVersion() {
	ctx := i.T().Context()
	now := time.Now().Unix()
	err := i.db.WithContext(ctx).
		Create(&dao.InvocationConfig{
			ID:          1000,
			Name:        "job_llm_invoke",
			BizID:       1,
			Description: "高端面试的配置",
			Ctime:       now,
			Utime:       now,
		}).Error
	require.NoError(i.T(), err)
	attrs := map[string]any{
		"engineer": map[string]any{
			"advanced": "需要有高并发高可用的工作经历",
			"basic":    "是个人就行",
		},
		"engineer1": map[string]any{
			"advanced": "随便搞搞就行",
			"basic":    "随便试试",
		},
	}
	attrsStr, err := json.Marshal(attrs)
	require.NoError(i.T(), err)

	err = i.db.WithContext(ctx).Create(&dao.InvocationConfigVersion{
		ID:      1199,
		InvID:   1000,
		ModelID: 2,
		Version: "v1.0.1",
		Prompt: `这是我的评判标准，你需要对用户的简历信息进行一个评判，而后输出
1. 如果简历信息达到要求，则执行下一个 llm_invoke
2. 否则，执行ask_user，要求用户补充更多
评判标准
{{ replace "${level}" .data.level "{{ .attr.${level} }}" }}
`,
		SystemPrompt: "",
		Attributes: sql.Null[string]{
			Valid: true,
			V:     string(attrsStr),
		},
		Status: domain.InvocationCfgVersionStatusActive.String(),
		Ctime:  now,
		Utime:  now,
	}).Error
	require.NoError(i.T(), err)
}

func (i *InvokeLLmFcall) TearDownSuite() {
	err := i.db.WithContext(i.T().Context()).
		Model(&dao.InvocationConfigVersion{}).
		Where("id = ?", 1199).Delete(&dao.InvocationConfigVersion{}).Error
	require.NoError(i.T(), err)
	err = i.db.WithContext(i.T().Context()).
		Model(&dao.InvocationConfig{}).
		Where("id = ?", 1000).Delete(&dao.InvocationConfig{}).Error
	require.NoError(i.T(), err)
}
