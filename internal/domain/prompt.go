package domain

import (
	"time"
)

type OwnerType string

func (o OwnerType) String() string {
	return string(o)
}

const (
	Personal     OwnerType = "personal"
	Organization OwnerType = "organization"
)

type Prompt struct {
	ID          int64
	Name        string
	Owner       int64
	OwnerType   OwnerType
	Description string
	// 当前发布版本的 id
	ActiveVersion int64
	// prompt 所有的版本信息
	Versions []PromptVersion
	Ctime    time.Time
	Utime    time.Time
}

type PromptVersion struct {
	ID            int64
	Label         string
	Content       string
	SystemContent string
	Temperature   float32
	TopN          float32
	MaxTokens     int
	Status        uint8
	Ctime         time.Time
	Utime         time.Time
}
