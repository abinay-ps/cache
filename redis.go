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

// This function will return a new Redis Cache Client
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

// This function will set a Key-value pair in Redis Cache
func SetRedisKey[T any](r *RedisClient, ctx context.Context, key string, value *T) {
	json, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	if r.Client != nil {
		r.Client.Set(ctx, key, string(json), 60*time.Second)
	}
}

// This function will fetch the value for a key from Redis Cache
func GetRedisKey[T any](r *RedisClient, ctx context.Context, key string) *T {
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

// This function will delete a key-value pair from Redis Cache
func DeleteRedisKey(r *RedisClient, ctx context.Context, key string) {
	if r.Client != nil {
		r.Client.Del(ctx, key)
	}
}

// This function will flush all Key-Value pairs from all indexes of a Redis Cache.
func FlushAllRedisDBs(r *RedisClient, ctx context.Context) {
	if r.Client != nil {
		r.Client.FlushAll(ctx)
	}
}

// This function will flush all Key-Value pairs of a particular index of a Redis Cache.
func FlushCurrentRedisDB(r *RedisClient, ctx context.Context) {
	if r.Client != nil {
		r.Client.FlushDB(ctx)
	}
}
