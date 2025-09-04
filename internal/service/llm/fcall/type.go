package fcall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrArgNotFound = errors.New("参数没有找到")
)

const (
	NameAskUser   = "ask_user"
	NameEmitJSON  = "emit_json"
	NameInvokeLLM = "invoke_llm"
)

//go:generate mockgen -source=./type.go -package=mocks -destination=./mocks/fcall.mock.go -typed FunctionCall
type FunctionCall interface {
	Name() string
	Call(ctx *Context, req Request) (Response, error)
}

type Context struct {
	context.Context
	// 根据发起LLM调用时传入的JSON Schema和用户输入的业务数据，经由LLM提取的结构化JSON数据
	// 由LLM 发起的 emit_json 函数调用设置
	JSONData map[string]any
	// Attachments 是每个functionCall的产物，每个functionCall如果想要，其他人共享都可以放在这个字段里
	Attachments map[string]string
}

func (c *Context) SetAttachment(key, val string) {
	if c.Attachments == nil {
		c.Attachments = map[string]string{}
	}
	c.Attachments[key] = val
}

func (c *Context) GetAttachment(key string) string {
	if c.Attachments == nil {
		return ""
	}
	return c.Attachments[key]
}

type Request struct {
	// 从 模型返回的数据来说，应该是一个 JSON 或者是一个 map，可能不同的模型也有差别
	// Args 是指 function call 里面传递回来的参数
	Args []byte
}

// GetArgs 获取序列化好的Args
func (r Request) GetArgs() (map[string]string, error) {
	args := make(map[string]string)
	err := json.Unmarshal(r.Args, &args)
	if err != nil {
		return nil, fmt.Errorf("序列化失败 %w", err)
	}
	return args, nil
}

// GetArg 获取具体参数的值
func (r Request) GetArg(key string) (string, error) {
	argMap, err := r.GetArgs()
	if err != nil {
		return "", err
	}
	val, ok := argMap[key]
	if !ok {
		return "", ErrArgNotFound
	}
	return val, nil
}

// Response 需要什么字段也不确定，按需要添加
type Response struct {
	Output string
	Status string // "in_progress", "completed", "incomplete"
}

func newCallErr(funcCall FunctionCall, err error) error {
	return fmt.Errorf("fcall: %s 发送错误 %w", funcCall.Name(), err)
}
