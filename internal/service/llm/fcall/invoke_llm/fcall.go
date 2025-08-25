package invoke_llm

import (
	"fmt"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm"
	"github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"
	"strconv"
)

const (
	invocationIDKey = "invocation_id"
	gjsonKey        = "gjson"
)

type FCall struct {
	svc  llm.Handler
	repo repository.InvocationConfigRepo
}

func (c *FCall) Name() string {
	return "invoke_llm"
}

func (c *FCall) Call(fctx *fcall.Context, req fcall.Request) (fcall.Response, error) {
	// 获取invokeconfig
	invocationVersion, err := c.getInvocationCfg(fctx, req)
	if err != nil {
		return fcall.Response{}, err
	}
	//

}

func (c *FCall) getInvocationCfg(fctx *fcall.Context, req fcall.Request) (domain.InvocationConfigVersion, error) {
	invocationIDStr, err := req.GetVal(invocationIDKey)
	if err != nil {
		return domain.InvocationConfigVersion{}, fcall.NewFcallErr(c, err)
	}
	invocationID, err := strconv.ParseInt(invocationIDStr, 10, 64)
	if err != nil {
		return domain.InvocationConfigVersion{}, fcall.NewFcallErr(c, fmt.Errorf("invocationID 数据类型不正确%w", err))
	}
	version, err := c.repo.GetActiveVersionByID(fctx, invocationID)
	if err != nil {
		return domain.InvocationConfigVersion{}, fcall.NewFcallErr(c, fmt.Errorf("未找到相关配置，配置id%d，err：%w", invocationID, err))
	}
	return version, nil
}

func (c *FCall) getgjsonExpr(fctx *fcall.Context, req fcall.Request) (string, error) {
	gjsonStr, err := req.GetVal(gjsonKey)
	if err != nil {
		return "", fcall.NewFcallErr(c, err)
	}
	return gjsonStr, nil
}
