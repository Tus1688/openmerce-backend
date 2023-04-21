package database

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

/*
0 for awaiting email verification (email, verification code) (ttl: 15 minutes)
1 for refresh token customer key: refresh_token  value: JSON of user-agent and id (ttl: 14 day??)
2 for refresh token staff key: refresh_token  value: JSON of user-agent, id, username, fin_user, inv_user, sys_admin (ttl: 14 day??)
*/
var RedisInstance []*redis.Client
var ctx = context.Background()

func NewRedis() error {
	// loop until 5 times
	for i := 0; i < 3; i++ {
		// create new redis client
		addr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
		client := redis.NewClient(&redis.Options{
			Addr:     os.Getenv(addr),
			Password: os.Getenv("REDIS_PASS"),
			DB:       i,
		})
		// ping redis
		_, err := client.Ping(ctx).Result()
		if err != nil {
			return err
		}
		// if ping success, add to redis instance
		RedisInstance = append(RedisInstance, client)
	}
	return nil
}
