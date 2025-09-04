package fcall

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/pkg/template"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

const (
	invocationIDKey = "invocation_id"
	gjsonKey        = "gjson_expr"
)

type InvokeLLMFuncCall struct {
	repo         *repository.InvocationConfigRepo
	renderClient *template.DefaultRender
}

func NewInvokeLLMFuncCall(repo *repository.InvocationConfigRepo,
	renderClient *template.DefaultRender,
) *InvokeLLMFuncCall {
	return &InvokeLLMFuncCall{
		repo:         repo,
		renderClient: renderClient,
	}
}

func (c *InvokeLLMFuncCall) Name() string {
	return NameInvokeLLM
}

func (c *InvokeLLMFuncCall) Call(ctx *Context, req Request) (Response, error) {
	invocationConfig, err := c.getInvocationConfig(ctx, req)
	if err != nil {
		return Response{}, newCallErr(c, err)
	}
	gjsonExpr, err := req.GetArg(gjsonKey)
	if err != nil {
		return Response{}, newCallErr(c, err)
	}
	prompt, err := c.getRenderedPrompt(ctx, gjsonExpr, invocationConfig)
	if err != nil {
		return Response{}, newCallErr(c, err)
	}
	// 调用llm接口
	// resp, err := c.svc.Stream(ctx, domain.ChatStreamRequest{
	// 	Sn: uuid.New().String(),
	// 	Messages: []domain.Message{
	// 		{
	// 			Role:    domain.USER,
	// 			Content: prompt,
	// 		},
	// 	},
	// 	InvocationConfigID: 0,
	// 	Uid:                0,
	// 	Key:                "",
	// 	PreviousResponseID: "",
	// })
	// if err != nil {
	// 	return Response{}, newCallErr(c, err)
	// }
	msgBytes, err := json.Marshal(&domain.Message{
		Role:    domain.USER,
		Content: prompt,
	})
	if err != nil {
		return Response{}, newCallErr(c, err)
	}
	ctx.SetAttachment(c.Name(), string(msgBytes))
	return Response{
		Output: "success",
		Status: "completed",
	}, nil
}

func (c *InvokeLLMFuncCall) getInvocationConfig(ctx *Context, req Request) (domain.InvocationConfigVersion, error) {
	invocationIDStr, err := req.GetArg(invocationIDKey)
	if err != nil {
		return domain.InvocationConfigVersion{}, err
	}
	invocationID, err := strconv.ParseInt(invocationIDStr, 10, 64)
	if err != nil {
		return domain.InvocationConfigVersion{}, fmt.Errorf("invocationID 数据类型不正确：%w", err)
	}
	version, err := c.repo.GetActiveVersionByID(ctx, invocationID)
	if err != nil {
		return domain.InvocationConfigVersion{}, fmt.Errorf("未找到相关配置，配置id=%d，err：%w", invocationID, err)
	}
	return version, nil
}

func (c *InvokeLLMFuncCall) getRenderedPrompt(ctx *Context, gjsonExpr string, version domain.InvocationConfigVersion) (string, error) {
	tplCtx := template.NewContext(ctx)
	err := tplCtx.SetVariable(template.NewVariable("data", ctx.JSONData))
	if err != nil {
		return "", err
	}
	attrs, err := version.Attributes.Get(gjsonExpr)
	if err != nil {
		return "", err
	}
	err = tplCtx.SetVariable(template.NewVariable("attr", attrs))
	if err != nil {
		return "", err
	}
	// 第一次渲染将所有的变量替换掉
	prompt, err := c.renderClient.Render(tplCtx, version.Prompt)
	if err != nil {
		return "", err
	}
	// 第二次渲染全部完毕
	return c.renderClient.Render(tplCtx, prompt)
}
