package invoke_llm

import (
	"encoding/json"
	"fmt"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/pkg/template"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"
	"github.com/google/uuid"
	"strconv"
)

const (
	invocationIDKey = "invocation_id"
	gjsonKey        = "gjson"
	invokeLLm       = "invoke_llm"
)

type FCall struct {
	svc          *service.ChatService
	repo         *repository.InvocationConfigRepo
	renderClient *template.DefaultRender
}

func NewFcall(svc *service.ChatService,
	renderClient *template.DefaultRender,
	repo *repository.InvocationConfigRepo) *FCall {
	return &FCall{
		svc:          svc,
		repo:         repo,
		renderClient: renderClient,
	}
}

func (c *FCall) Name() string {
	return "invoke_llm"
}

func (c *FCall) Call(fctx *fcall.Context, req fcall.Request) (fcall.Response, error) {
	invocationVersion, err := c.getInvocationCfg(fctx, req)
	if err != nil {
		return fcall.Response{}, fcall.NewFcallErr(c, err)
	}
	gjsonExpr := c.getGjsonExpr(req)

	prompt, err := c.render(fctx, fctx.JSONData, gjsonExpr, invocationVersion)
	if err != nil {
		return fcall.Response{}, fcall.NewFcallErr(c, err)
	}
	// 调用llm接口
	sn := uuid.New().String()
	resp, err := c.svc.Chat(fctx, sn, []domain.Message{
		{
			Role:    domain.USER,
			Content: prompt,
		},
	})
	if err != nil {
		return fcall.Response{}, fcall.NewFcallErr(c, err)
	}
	llmResp, err := json.Marshal(resp)
	if err != nil {
		return fcall.Response{}, fcall.NewFcallErr(c, err)
	}
	fctx.Attachments[invokeLLm] = string(llmResp)
	return fcall.Response{}, nil
}

func (c *FCall) render(fctx *fcall.Context,
	args map[string]any,
	gjsonExpr string,
	version domain.InvocationConfigVersion) (string, error) {
	ctx := template.NewContext(fctx)
	err := ctx.SetVariable(template.NewVariable("data", args))
	if err != nil {
		return "", err
	}
	attrs := version.Attributes.GetAttribute(gjsonExpr)
	err = ctx.SetVariable(template.NewVariable("attr", attrs))
	if err != nil {
		return "", err
	}
	// 第一次渲染将所有的变量替换掉
	prompt, err := c.renderClient.Render(ctx, version.Prompt)
	if err != nil {
		return "", err
	}
	// 第二次渲染全部完毕
	return c.renderClient.Render(ctx, prompt)
}

func (c *FCall) getInvocationCfg(fctx *fcall.Context, req fcall.Request) (domain.InvocationConfigVersion, error) {
	invocationIDStr, err := req.GetVal(invocationIDKey)
	if err != nil {
		return domain.InvocationConfigVersion{}, err
	}
	invocationID, err := strconv.ParseInt(invocationIDStr, 10, 64)
	if err != nil {
		return domain.InvocationConfigVersion{}, fmt.Errorf("invocationID 数据类型不正确%w", err)
	}
	version, err := c.repo.GetActiveVersionByID(fctx, invocationID)
	if err != nil {
		return domain.InvocationConfigVersion{}, fmt.Errorf("未找到相关配置，配置id%d，err：%w", invocationID, err)
	}
	return version, nil
}

func (c *FCall) getGjsonExpr(req fcall.Request) string {
	val, _ := req.GetVal(gjsonKey)
	return val
}
