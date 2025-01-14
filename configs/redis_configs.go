package configs

import (
	"github.com/go-redis/redis/v8"
)

func ConnectRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 서버 주소
		Password: "",               // Redis 비밀번호 (설정되어 있지 않으면 빈 문자열)
		DB:       0,                // 기본 DB (0번 DB)
	})

	return client
}
