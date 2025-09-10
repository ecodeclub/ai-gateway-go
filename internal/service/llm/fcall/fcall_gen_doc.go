package fcall

import (
	"github.com/goccy/go-json"
)

type GenDocFunctionCall struct{}

func NewGenDocFunctionCall() *GenDocFunctionCall {
	return &GenDocFunctionCall{}
}

func (e *GenDocFunctionCall) Name() string {
	return NameGenDoc
}

const jsonDataNameDocs = "doc"

func (e *GenDocFunctionCall) Call(ctx *Context, req Request) (Response, error) {
	d, err := req.GetArg(jsonDataNameDocs)
	if err != nil {
		return Response{}, err
	}
	var doc Doc
	err = json.Unmarshal([]byte(d), &doc)
	if err != nil {
		return Response{}, err
	}
	var content string
	switch doc.Type {
	case DocMarkDown:
		content = doc.Content
	case DocPDF:
		//todo: 将对应的 markdown 格式渲染成为对应的 pdf
	default:
	}
	ctx.SetAttachment(e.Name(), content)
	return Response{}, nil
}

type Doc struct {
	Content string  `json:"content"`
	Type    DocType `json:"type"`
}

type DocType string

const (
	DocMarkDown = "markdown"
	DocPDF      = "pdf"
)
