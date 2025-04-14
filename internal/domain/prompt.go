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
	Content     string
	Description string
	Ctime       time.Time
	Utime       time.Time
}
