package domain

import (
	"github.com/ecodeclub/ekit"
)

const (
	UNKNOWN = iota
	USER
	ASSISTANT
	SYSTEM
	TOOL
)

type Conversation struct {
	Sn       string
	Messages []Message
	time     string
}
type Message struct {
	ID               int64
	Role             int64
	Content          string
	ReasoningContent string
}

type ChatResponse struct {
	Sn       string
	Response Message
	Metadata ekit.AnyValue
}
