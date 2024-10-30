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
)

type API struct {
	*sturdyc.Client[string]
	RedisClient   *RedisClient
	DatabaseCalls uint64
	RedisCalls    uint64
}

func NewAPI(addr string, password string, index int, capacity int, shards int, batchSize int,
	batchBufferTimeout time.Duration, evictionPercentage int, maxRefreshDelay time.Duration,
	minRefreshDelay time.Duration, retryBaseDelay time.Duration, ttl time.Duration) (*API, error) {

	c := NewCacheClient(capacity, shards, batchSize, batchBufferTimeout, evictionPercentage,
		maxRefreshDelay, minRefreshDelay, retryBaseDelay, ttl)

	redisClient := NewRedisClient(addr, password, index)
	if redisClient.Client == nil {
		return &API{c, redisClient, 0, 0}, errors.New("error initializing redis client")
	}

	return &API{c, redisClient, 0, 0}, nil
}

func GetVal[T any](a *API, ctx context.Context, key string) (*T, error) {
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

func fetchFromRedis[T any](a *API, ctx context.Context, key string) *T {
	atomic.AddUint64(&a.RedisCalls, 1)
	if a.RedisClient == nil {
		return nil
	}
	descriptions := GetRedisKey[T](a.RedisClient, ctx, key)
	return descriptions
}

func Fetch[T any](api *API, key string) (*T, error) {
	var wg sync.WaitGroup
	var values *T
	var err error
	wg.Add(1)
	go func() {
		values, err = GetVal[T](api, context.Background(), key)
		wg.Done()
	}()
	wg.Wait()
	return values, err
}

type Handler interface {
	GetAPI() *API
}

func ReturnNilOrZero[T any]() T {
	var result T
	if reflect.ValueOf(result).IsNil() {
		return result
	}
	return result
}

func FetchData[T any](handler Handler, rserver string, rpwd string, rindex int, key string, fn any, args ...any) (T, error) {
	if handler.GetAPI().RedisClient.Client == nil {
		redisClient := NewRedisClient(rserver, rpwd, rindex)
		if redisClient != nil {
			handler.GetAPI().RedisClient = redisClient
		}
	}
	values, _ := Fetch[T](handler.GetAPI(), key)
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
		if handler.GetAPI().RedisClient != nil {
			SetRedisKey[T](handler.GetAPI().RedisClient, context.Background(), key, zero)
		}
		SetLocalKey(handler.GetAPI().Client, context.Background(), key, zero)
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
		if handler.GetAPI().RedisClient != nil {
			SetRedisKey[T](handler.GetAPI().RedisClient, context.Background(), key, &value)
		}
		SetLocalKey(handler.GetAPI().Client, context.Background(), key, &value)
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
		if handler.GetAPI().RedisClient != nil {
			SetRedisKey[T](handler.GetAPI().RedisClient, context.Background(), key, &value)
		}
		SetLocalKey(handler.GetAPI().Client, context.Background(), key, &value)
		return value, nil
	}
	return ReturnNilOrZero[T](), nil
}
