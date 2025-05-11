package infra

import (
	"github.com/ecodeclub/ginx/session"
	redis2 "github.com/ecodeclub/ginx/session/redis"
	"time"
)

func Init() {
	provider := redis2.NewSessionProvider(InitRedis(), "VGhpcyBpcyBhIHNlY3JldCB0aGF0IG5vYm9keSBjYW4gZ3Vlc3M=", time.Hour)
	session.SetDefaultProvider(provider)
}
