# Cache Library for IT 2.0 application
This library aims to utilize Local Cache and Redis Cache for caching the responses for better performance of IT 2.0 applications.

## Installation
To install Cache Library, run:

```bash
go get github.com/abinay-ps/cache
```

## Usage

### Importing the Library

```go
import "github.com/abinay-ps/cache"
```

## Index
```go
func DeleteLocalKey(c *sturdyc.Client[string], ctx context.Context, key string)
func DeleteRedisKey(r *RedisClient, ctx context.Context, key string)
func Fetch[T any](api *API, key string) (*T, error)
func FetchData[T any](handler Handler, key string, fn any, args ...any) (T, error)
func FlushAllRedisDBs(r *RedisClient, ctx context.Context)
func FlushCurrentRedisDB(r *RedisClient, ctx context.Context)
func GetRedisKey[T any](r *RedisClient, ctx context.Context, key string) *T
func GetVal[T any](a *API, ctx context.Context, key string) (*T, error)
func NewCacheClient(capacity int, shards int, batchSize int, batchBufferTimeout time.Duration, evictionPercentage int, maxRefreshDelay time.Duration, minRefreshDelay time.Duration, retryBaseDelay time.Duration, ttl time.Duration) *sturdyc.Client[string]
func ReturnNilOrZero[T any]() T
func SetLocalKey[T any](c *sturdyc.Client[string], ctx context.Context, key string, value *T)
func SetRedisKey[T any](r *RedisClient, ctx context.Context, key string, value *T)
func NewAPI(addr string, password string, index int, capacity int, shards int, batchSize int, batchBufferTimeout time.Duration, evictionPercentage int, maxRefreshDelay time.Duration, minRefreshDelay time.Duration, retryBaseDelay time.Duration, ttl time.Duration) (*API, error)
func NewRedisClient(addr string, password string, index int) *RedisClient
```

## Interface: type Handler
```go
type Handler interface {
    GetAPI() *API
}
```

## Struct Definition: type API
This struct variable will hold Local Cache and Redis Cache Clients.
```go
type API struct {
    *sturdyc.Client[string]
    RedisClient   *RedisClient
    DatabaseCalls uint64
    RedisCalls    uint64
    RedisServer   string
    RedisPassword string
    RedisDBIndex  int
}
```

## Functions
## func DeleteLocalKey
```go
func DeleteLocalKey(c *sturdyc.Client[string], ctx context.Context, key string)
```
This function will delete a key-value pair in Local Cache.

## func DeleteRedisKey
```go
func DeleteRedisKey(r *RedisClient, ctx context.Context, key string)
```
This function will delete a key-value pair from Redis Cache

## func FetchData
```go
func FetchData[T any](handler Handler, key string, fn any, args ...any) (T, error)
```
This function will follow Lazy Loading Principle where it will first check the key in local cache and if not found it will check Redis cache. If not found, then it will fall back to database

## func FlushAllRedisDBs
```go
func FlushAllRedisDBs(r *RedisClient, ctx context.Context)
```
This function will flush all Key-Value pairs from all indexes of a Redis Cache.

## func FlushCurrentRedisDB
```go
func FlushCurrentRedisDB(r *RedisClient, ctx context.Context)
```
This function will flush all Key-Value pairs of a particular index of a Redis Cache.

## func GetRedisKey
```go
func GetRedisKey[T any](r *RedisClient, ctx context.Context, key string) *T
```
This function will fetch the value for a key from Redis Cache

## func GetVal
```go
func GetVal[T any](a *API, ctx context.Context, key string) (*T, error)
```
This function will first check the key in local cache. If found, it will return the value and if not found, it will check Redis cache. If found, will set the key and its value in local cache and returns it. Otherwise, it will return nil.

## func NewCacheClient
```go
func NewCacheClient(capacity int, shards int, batchSize int, batchBufferTimeout time.Duration,
    evictionPercentage int, maxRefreshDelay time.Duration, minRefreshDelay time.Duration,
    retryBaseDelay time.Duration, ttl time.Duration) *sturdyc.Client[string]
```
This function will return a new Local Cache Client.

## func SetLocalKey
```go
func SetLocalKey[T any](c *sturdyc.Client[string], ctx context.Context, key string, value *T)
```
This function will set a key-value pair in Local Cache.

## func SetRedisKey
```go
func SetRedisKey[T any](r *RedisClient, ctx context.Context, key string, value *T)
```
This function will set a Key-value pair in Redis Cache

## func NewAPI
```go
func NewAPI(addr string, password string, index int, capacity int, shards int, batchSize int,
    batchBufferTimeout time.Duration, evictionPercentage int, maxRefreshDelay time.Duration,
    minRefreshDelay time.Duration, retryBaseDelay time.Duration, ttl time.Duration) (*API, error)
```
This function will initiate a new Cache Client and a new Redis Client

## func NewRedisClient
```go
func NewRedisClient(addr string, password string, index int) *RedisClient
```
This function will return a new Redis Cache Client