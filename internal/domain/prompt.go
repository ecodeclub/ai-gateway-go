package domain

import (
	"time"
)

type Prompt struct {
	ID          int64
	Name        string
	Biz         string
	Pattern     string
	Description string
	Ctime       time.Time
	Utime       time.Time
}
