package infra

import (
	"github.com/ecodeclub/ginx/session"
	redis2 "github.com/ecodeclub/ginx/session/redis"
	"time"
)

// Init 初始化会话提供者
// 使用 Redis 创建一个新的会话提供者并设置为默认提供者
// 密钥用于加密会话数据，过期时间设置为 1 小时
func Init() {
	provider := redis2.NewSessionProvider(InitRedis(), "VGhpcyBpcyBhIHNlY3JldCB0aGF0IG5vYm9keSBjYW4gZ3Vlc3M=", time.Hour)
	session.SetDefaultProvider(provider)
}
