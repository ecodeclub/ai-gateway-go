package fcall

const jsonDataNameQuestions = "questions"

type AskUserFunctionCall struct {
}

func NewAskUserFunctionCall() *AskUserFunctionCall {
	return &AskUserFunctionCall{}
}

func (a *AskUserFunctionCall) Name() string {
	return NameAskUser
}

func (a *AskUserFunctionCall) Call(ctx *Context, req Request) (Response, error) {
	val, err := req.GetArg(jsonDataNameQuestions)
	if err != nil {
		return Response{}, err
	}
	ctx.SetAttachment(a.Name(), val)
	return Response{
		Output: "请稍等，用户的回答将在下一次请求中.....",
		Status: "in_progress",
	}, nil
}
