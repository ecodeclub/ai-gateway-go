package infra

import (
	"time"

	"github.com/ecodeclub/ginx/session"
	redis2 "github.com/ecodeclub/ginx/session/redis"
)

func Init() {
	provider := redis2.NewSessionProvider(InitRedis(), "VGhpcyBpcyBhIHNlY3JldCB0aGF0IG5vYm9keSBjYW4gZ3Vlc3M=", time.Hour)
	session.SetDefaultProvider(provider)
}
