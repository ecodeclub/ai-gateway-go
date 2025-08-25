package invoke_llm

import "github.com/ecodeclub/ai-gateway-go/internal/service/llm/fcall"

type FCall struct {

}

func (F *FCall) Name() string {
	return "invoke_llm"
}

func (F *FCall) Call(fctx *fcall.Context, req fcall.Request) (fcall.Response, error) {

}

