package fcall

import (
	"errors"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/pkg/jsonx"
)

var (
	ErrArgNotFound = errors.New("参数没有找到")
)

// FunctionCall 被认为是一个 Step
// 也就是一个 function call 本质上就可以看做是修改 ctx
// 1. 修改 ctx 中的公共字段
// 2. 在 ctx 中创建一个新的 Step，并且将自己的数据放入到 Step 中
// 3. 如果某个后续步骤依赖前置的 Step，那么需要自己从 Step 中查找
//
//go:generate mockgen -source=./type.go -package=mocks -destination=./mocks/fcall.mock.go -typed FunctionCall
type FunctionCall interface {
	Name() string
	Call(ctx *domain.StreamContext, req Request) (Response, error)
}

type Request struct {
	// 从 模型返回的数据来说，应该是一个 JSON 或者是一个 map，可能不同的模型也有差别
	// Args 是指 function call 里面传递回来的参数
	Args jsonx.RawJSON
}

// Response 需要什么字段也不确定，按需要添加
// 用作未来的扩展点
type Response struct {
	// 返回给大模型的数据，目前只考虑返回 string
	Content string

	// 需要执行下一步调用
	NextInvCfgID int64
}
