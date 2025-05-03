package domain

import "time"

type BizConfig struct {
	ID        int64
	OwnerID   int64
	OwnerType string // "person" or "organization"
	Config    string // JSON string
	Ctime     time.Time
	Utime     time.Time
}
