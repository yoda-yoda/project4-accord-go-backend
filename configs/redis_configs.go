package configs

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func GetRedisClient() *redis.Client {
	return RedisClient
}

var Ctx = context.Background()
