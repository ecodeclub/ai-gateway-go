package web

import (
	"github.com/ecodeclub/ai-gateway-go/internal/errs"
	"github.com/ecodeclub/ginx"
)

var (
	systemErrorResult = ginx.Result{
		Code: errs.SystemError.Code,
		Msg:  errs.SystemError.Msg,
	}
)
