package fcall

type FunctionCall interface {
	Name() string
	Call(ctx *Context, req Request) (Response, error)
}
type Context struct {
	// 这里的data暂定是一个map，然后每一个fcall如果需要将处理完的数据交给下游fcall，都可以存储在Data中
	// 例如 Data => map[string]any{
	//	"emit_json": []byte(`{}`)
	//}
	Data map[string]any
	// functioncall 会在这里填入数据，会原样返回调用者 调用者是指客户端, 调用什么方式去解析
	Attachments map[string]string
}

type Request struct {
	// 从 模型返回的数据来说，应该是一个 JSON 或者是一个 map，可能不同的模型也有差别
	// Args 是指 function call 里面传递回来的参数
	Args []byte
}

// 需要什么字段也不确定，按需要添加
type Response struct {
}
