package domain

type BizConfig struct {
	ID        int64
	OwnerID   int64
	OwnerType string // "person" or "organization"
	Token     string // 加密存储
	Config    string // JSON string
	CreatedAt int64
	UpdatedAt int64
}
