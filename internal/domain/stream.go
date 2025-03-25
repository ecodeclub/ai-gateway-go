package domain

type StreamRequest struct {
	Id   string
	Text string
}
type StreamEvent struct {
	Content string
	Done    bool
	Error   error
}
