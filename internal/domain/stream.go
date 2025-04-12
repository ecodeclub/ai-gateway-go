package domain

type LLMRequest struct {
	Id   string
	Text string
}
type StreamEvent struct {
	ReasoningContent string
	Content          string
	Done             bool
	Error            error
}

type LLMResponse struct {
	ReasoningContent string
	Content          string
}
