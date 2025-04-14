package domain

import (
	"time"
)

type BizConfig struct {
	ID        string
	OwnerID   int64
	OwnerType string // "person" or "organization"
	Token     string // 加密存储
	Config    string // JSON string
	Quota     int64
	UsedQuota int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
