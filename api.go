package cache

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/creativecreature/sturdyc"
)

type API struct {
	*sturdyc.Client[string]
	RedisClient   *RedisClient
	DatabaseCalls uint64
	RedisCalls    uint64
}

func NewAPI(addr string, password string, index int) (*API, error) {

	c := NewCacheClient()

	redisClient := NewRedisClient(addr, password, index)
	if redisClient.Client == nil {
		return &API{c, redisClient, 0, 0}, errors.New("error initializing redis client")
	}

	return &API{c, redisClient, 0, 0}, nil
}

func GetVal[T any](a *API, ctx context.Context, key string) (*T, error) {
	// count := false

	fetchFn := func(ctx context.Context) (string, error) {
		// count = true
		// fmt.Println("Inside fetchFn function")
		if a.RedisClient.Client != nil {
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
	descriptions := Get[T](a.RedisClient, ctx, key)
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
