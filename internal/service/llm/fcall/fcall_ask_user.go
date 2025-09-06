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
	return Response{}, nil
}
