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

type FunctionCall interface {
	Name() string
	Call(fctx *Context, req Request) (Response, error)
}

type Context struct {
	context.Context
	// 这里的是用户输入的数据
	JSONData string
	// Attachments 是每个functionCall的产物，每个functioncall如果想要，其他人共享都可以放在这个字段里，健是functionCall的name。
	Attachments map[string]string
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
	return args, fmt.Errorf("序列化失败 %w", err)
}

// GetVal 获取具体的值
func (r Request) GetVal(key string) (string, error) {
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

// 需要什么字段也不确定，按需要添加
type Response struct {
}

func NewFcallErr(fcall FunctionCall, err error) error {
	return fmt.Errorf("functionCall: %s 发送错误 %w", fcall.Name(), err)
}
