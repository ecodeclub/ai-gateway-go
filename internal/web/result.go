package web

import (
	"github.com/ecodeclub/ai-gateway-go/internal/errs"
	"github.com/ecodeclub/ginx"
)

// systemErrorResult 定义系统错误时返回的标准结果
// 包含错误码和错误信息，用于统一处理服务器内部错误
var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}
)
