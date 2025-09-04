package fcall

import (
	"encoding/json"
)

const jsonDataName = "data"

type EmitJsonFunctionCall struct {
}

func NewEmitJsonFunctionCall() *EmitJsonFunctionCall {
	return &EmitJsonFunctionCall{}
}

func (e *EmitJsonFunctionCall) Name() string {
	return NameEmitJSON
}

func (e *EmitJsonFunctionCall) Call(ctx *Context, req Request) (Response, error) {
	val, err := req.GetArg(jsonDataName)
	if err != nil {
		return Response{}, err
	}
	jsonData := make(map[string]any)
	err = json.Unmarshal([]byte(val), &jsonData)
	if err != nil {
		return Response{}, err
	}
	ctx.JSONData = jsonData
	return Response{
		Output: "success",
		Status: "completed",
	}, nil
}
