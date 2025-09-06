package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

func TestLLMFuncCall(t *testing.T) {
	suite.Run(t, new(TestLLMFuncCallSuite))
}

type TestLLMFuncCallSuite struct {
	suite.Suite
	db *gorm.DB
}

func (s *TestLLMFuncCallSuite) SetupSuite() {
	db := testioc.InitDB()
	s.db = db
}

// Call 在参数反序列化出错时应返回错误，并且不应写入附件
func (s *TestLLMFuncCallSuite) TestInvokeLLM_Call() {
	t := s.T()

	configID := int64(1000)
	versionID := int64(1199)

	t.Cleanup(func() {
		err := s.db.WithContext(context.Background()).
			Model(&dao.InvocationConfigVersion{}).
			Where("id = ?", versionID).Delete(&dao.InvocationConfigVersion{}).Error
		require.NoError(t, err)
		err = s.db.WithContext(context.Background()).
			Model(&dao.InvocationConfig{}).
			Where("id = ?", configID).Delete(&dao.InvocationConfig{}).Error
		require.NoError(t, err)
	})

	// 构造InvocationConfig 表记录
	now := time.Now().Unix()
	err := s.db.WithContext(t.Context()).
		Create(&dao.InvocationConfig{
			ID:          configID,
			Name:        "job_llm_invoke",
			BizID:       1,
			Description: "高端面试的配置",
			Ctime:       now,
			Utime:       now,
		}).Error
	require.NoError(t, err)

	// 构造 InvocationConfigVersion 表记录
	err = s.db.WithContext(t.Context()).Create(&dao.InvocationConfigVersion{
		ID:      versionID,
		InvID:   configID,
		ModelID: 2,
		Version: "v1.0.1",
		Prompt: `这是我的评判标准，你需要对用户的简历信息进行一个评判，而后输出
1. 如果简历信息达到要求，则执行下一个 llm_invoke
2. 否则，执行ask_user，要求用户补充更多
评判标准
{{ replace "${level}" .data.level "{{ .attr.${level} }}" }}
`,
		SystemPrompt: "",
		Attributes: func() sql.Null[string] {
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
			attrsStr, err1 := json.Marshal(attrs)
			require.NoError(t, err1)
			return sql.Null[string]{
				Valid: true,
				V:     string(attrsStr),
			}
		}(),
		Status: domain.InvocationCfgVersionStatusActive.String(),
		Ctime:  now,
		Utime:  now,
	}).Error
	require.NoError(t, err)

	// 构造mock Chat Handler
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 初始化APP
	app := testioc.InitApp(testioc.TestOnly{LLM: mocks.NewMockHandler(ctrl)})

	ctx := &fcall.Context{
		Context: t.Context(),
		JSONData: map[string]any{
			"level": "advanced",
		}}
	req := fcall.Request{Args: []byte(fmt.Sprintf(`{"invocation_id":"%d","gjson":"engineer"}`, configID))}

	// 执行调用大模型
	_, err = app.InvokeLLMFuncCall.Call(ctx, req)
	require.NoError(t, err)

	assert.Equal(t,
		ctx.Attachments,
		func() map[string]string {
			respByte, err1 := json.Marshal(&domain.Message{
				Role: domain.USER,
				Content: `这是我的评判标准，你需要对用户的简历信息进行一个评判，而后输出
1. 如果简历信息达到要求，则执行下一个 llm_invoke
2. 否则，执行ask_user，要求用户补充更多
评判标准
需要有高并发高可用的工作经历
`,
			})
			require.NoError(t, err1)
			return map[string]string{
				app.InvokeLLMFuncCall.Name(): string(respByte),
			}
		}(),
	)
}
