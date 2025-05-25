package infra

import (
	"github.com/redis/go-redis/v9"
)

// rdb 是全局的 Redis 命令执行器实例
// 用于在整个项目中执行 Redis 操作
var rdb redis.Cmdable

// InitRedis 初始化 Redis 客户端
// 如果已经存在一个客户端实例则直接返回，否则创建一个新的客户端实例
// 默认连接本地 Redis 服务（localhost:6379）
func InitRedis() redis.Cmdable {
	if rdb != nil {
		return rdb
	}
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}
