package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/creativecreature/sturdyc"
	"github.com/gin-gonic/gin"
)

// This struct variable will hold Local Cache and Redis Cache Clients.
type Cache struct {
	*sturdyc.Client[string]
	RedisClient   *RedisClient
	DatabaseCalls uint64
	RedisCalls    uint64
	RedisServer   string
	RedisPassword string
	RedisDBIndex  int
}

// This function will initiate a new Cache Client and a new Redis Client
func NewCache(addr string, password string, index int, capacity int, shards int, batchSize int,
	batchBufferTimeout time.Duration, evictionPercentage int, maxRefreshDelay time.Duration,
	minRefreshDelay time.Duration, retryBaseDelay time.Duration, ttl time.Duration) (*Cache, error) {

	c := NewCacheClient(capacity, shards, batchSize, batchBufferTimeout, evictionPercentage,
		maxRefreshDelay, minRefreshDelay, retryBaseDelay, ttl)

	redisClient := NewRedisClient(addr, password, index)
	if redisClient.Client == nil {
		return &Cache{c, redisClient, 0, 0, addr, password, index}, errors.New("error initializing redis client")
	}

	return &Cache{c, redisClient, 0, 0, addr, password, index}, nil
}

// This function will first check the key in local cache. If found, it will return the value and if not found,
// it will check Redis cache. If found, will set the key and its value in local cache and returns it. Otherwise, it will return nil.
func GetVal[T any](a *Cache, ctx context.Context, key string) (*T, error) {
	fetchFn := func(ctx context.Context) (string, error) {
		val := fetchFromRedis[T](a, ctx, key)
		if val == nil {
			return "", nil
		}
		jsonVal, err := json.Marshal(val)
		if err != nil {
			return "", err
		}
		return string(jsonVal), nil
	}
	strVal, err := a.GetOrFetch(ctx, key, fetchFn)
	if err != nil {
		return nil, err
	}
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

// This function will check Redis cache for the key and if found, will return the value. Otherwise, it will return nil.
func fetchFromRedis[T any](a *Cache, ctx context.Context, key string) *T {
	atomic.AddUint64(&a.RedisCalls, 1)
	if a.RedisClient == nil {
		return nil
	}
	descriptions := GetRedisKey[T](a.RedisClient, ctx, key)
	return descriptions
}

func Fetch[T any](cache *Cache, ctx *gin.Context, key string) (*T, error) {
	var wg sync.WaitGroup
	var values *T
	var err error
	wg.Add(1)
	go func() {
		values, err = GetVal[T](cache, ctx.Request.Context(), key)
		wg.Done()
	}()
	wg.Wait()
	return values, err
}

type Handler interface {
	GetCache() *Cache
}

func ReturnNilOrZero[T any]() T {
	var result T
	if reflect.ValueOf(result).IsNil() {
		return result
	}
	return result
}

// This function will follow Lazy Loading Principle where it will first check the key in local cache and if not found
// it will check Redis cache. If not found, then it will fall back to database
func FetchData[T any](ctx *gin.Context, handler Handler, key string, fn any, args ...any) (T, error) {
	if handler.GetCache().RedisClient.Client == nil {
		redisClient := NewRedisClient(handler.GetCache().RedisServer, handler.GetCache().RedisPassword, handler.GetCache().RedisDBIndex)
		if redisClient != nil {
			handler.GetCache().RedisClient = redisClient
		}
	}
	values, _ := Fetch[T](handler.GetCache(), ctx, key)
	if values != nil {
		return *values, nil
	}
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if len(args) != fnType.NumIn() {
		return ReturnNilOrZero[T](), fmt.Errorf("expected %d arguments, got %d while calling function %v", fnType.NumIn(), len(args), runtime.FuncForPC(fnValue.Pointer()).Name())
	}

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		argValue := reflect.ValueOf(arg)
		if argValue.Type() != fnType.In(i) {
			return ReturnNilOrZero[T](), fmt.Errorf("argument %d expected type %s, got %s while calling function %v", i, fnType.In(i), argValue.Type(), runtime.FuncForPC(fnValue.Pointer()).Name())
		}
		in[i] = argValue
	}

	results := fnValue.Call(in)
	if len(results) == 0 {
		var zero *T
		if handler.GetCache().RedisClient != nil {
			SetRedisKey[T](handler.GetCache().RedisClient, ctx.Request.Context(), key, zero)
		}
		SetLocalKey(handler.GetCache().Client, key, zero)
		return ReturnNilOrZero[T](), nil
	}

	if len(results) == 1 {
		if err, ok := results[0].Interface().(error); ok {
			return ReturnNilOrZero[T](), err
		}
		value, ok := results[0].Interface().(T)
		if !ok {
			return ReturnNilOrZero[T](), fmt.Errorf("expected return type %T, got %T", (*new(T)), results[0].Interface())
		}
		if handler.GetCache().RedisClient != nil {
			SetRedisKey[T](handler.GetCache().RedisClient, ctx.Request.Context(), key, &value)
		}
		SetLocalKey(handler.GetCache().Client, key, &value)
		return value, nil
	}
	if len(results) == 2 {
		value, ok := results[0].Interface().(T)
		if !ok {
			return ReturnNilOrZero[T](), fmt.Errorf("expected return type %T, got %T", (*new(T)), results[0].Interface())
		}
		if err, ok := results[1].Interface().(error); ok && err != nil {
			return ReturnNilOrZero[T](), err
		}
		if handler.GetCache().RedisClient != nil {
			SetRedisKey[T](handler.GetCache().RedisClient, ctx.Request.Context(), key, &value)
		}
		SetLocalKey(handler.GetCache().Client, key, &value)
		return value, nil
	}
	return ReturnNilOrZero[T](), nil
}
