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
3 for customer_cart counts (ttl: 14 day) key: customer_id value: counts
4 for area suggestion result for global (ttl: 30 day) key: area_id value: JSON of area response
5 for get rates by product result for global (ttl: 10 day): key: product_id_area_id value: JSON of freight response
6 for total sold by product for global (ttl: 10 day): key: product_id value: total sold
*/
var RedisInstance []*redis.Client
var ctx = context.Background()

func NewRedis() error {
	for i := 0; i < 7; i++ {
		// create new redis client
		addr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
		client := redis.NewClient(&redis.Options{
			Addr:     addr,
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
