package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(addr string, password string, index int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       index,
		Protocol: 3, //RESP 3

	})
	//Ping the connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return &RedisClient{Client: nil}
	}
	return &RedisClient{Client: rdb}
}

func Set[T any](r *RedisClient, ctx context.Context, key string, value *T) {
	json, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	if r.Client != nil {
		r.Client.Set(ctx, key, string(json), 60*time.Second)
	}
}

func Get[T any](r *RedisClient, ctx context.Context, key string) *T {
	if r.Client != nil {
		val, err := r.Client.Get(ctx, key).Result()
		if err != nil {
			return nil
		}
		var result T
		err = json.Unmarshal([]byte(val), &result)
		if err != nil {
			panic(err)
		}
		return &result
	}
	return nil
}

func Delete(r *RedisClient, ctx context.Context, key string) {
	if r.Client != nil {
		r.Client.Del(ctx, key)
	}
}
