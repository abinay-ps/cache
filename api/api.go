package api

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/abinay-ps/cache/cache"
	"github.com/abinay-ps/cache/redis"

	"github.com/creativecreature/sturdyc"
)

type API struct {
	*sturdyc.Client[string]
	RedisClient   *redis.RedisClient
	DatabaseCalls uint64
	RedisCalls    uint64
}

// func NewAPI(cfg config.Econfig) (*API, error) {

// 	c := cache.NewCacheClient()

// 	redisClient := redis.NewRedisClient(cfg)
// 	if redisClient.Client == nil {
// 		fmt.Println("Warning: Redis is unavailable, continuing without Redis cache.")
// 	}

// 	return &API{c, redisClient, 0, 0}, nil
// }

func NewAPI(addr string, password string, index int) (*API, error) {

	c := cache.NewCacheClient()

	redisClient := redis.NewRedisClient(addr, password, index)
	// if redisClient.Client == nil {
	// 	fmt.Println("Warning: Redis is unavailable, continuing without Redis cache.")
	// }

	return &API{c, redisClient, 0, 0}, nil
}

func GetVal[T any](a *API, ctx context.Context, key string) (*T, error) {
	// count := false

	fetchFn := func(ctx context.Context) (string, error) {
		// count = true
		// fmt.Println("Inside fetchFn function")
		if a.RedisClient != nil {
			val, err := fetchFromRedis[T](a, ctx, key)
			if err != nil {
				return "", err
			}
			if val == nil {
				return "", err
			}
			jsonVal, err := json.Marshal(val)
			if err != nil {
				return "", err
			}
			return string(jsonVal), nil
		}
		return "", nil
	}

	strVal, err := a.GetOrFetch(ctx, key, fetchFn)
	if err != nil {
		return nil, err
	}
	// if !count {
	// 	fmt.Println("Data from Local Cache")
	// }
	if strVal == "" {
		a.Delete(key)
		return nil, nil
	}

	var result T
	err = json.Unmarshal([]byte(strVal), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func fetchFromRedis[T any](a *API, ctx context.Context, key string) (*T, error) {
	atomic.AddUint64(&a.RedisCalls, 1)
	// fmt.Println("Inside fetchFromDatabase1 function")
	if a.RedisClient == nil {
		return nil, nil
	}
	descriptions := redis.Get[T](a.RedisClient, ctx, key)
	// if descriptions != nil {
	// 	fmt.Println("Data from Redis")
	// }
	return descriptions, nil
}

func Fetch[T any](api *API, key string) (*T, error) {
	var wg sync.WaitGroup
	var values *T
	var err error

	wg.Add(1)
	go func() {
		// startTime := time.Now()
		values, err = GetVal[T](api, context.Background(), key)
		// elapsedTime := time.Since(startTime).Milliseconds()
		//log.Printf("got values: %v\n", values)
		// log.Printf("Time taken to fetch: %v ms\n", elapsedTime)
		wg.Done()
	}()

	wg.Wait()
	// log.Printf("Total Redis calls: %d\n", api.RedisCalls)
	return values, err
}
