package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/creativecreature/sturdyc"
)

func NewCacheClient(capacity int, shards int, batchSize int, batchBufferTimeout time.Duration,
	evictionPercentage int, maxRefreshDelay time.Duration, minRefreshDelay time.Duration,
	retryBaseDelay time.Duration, ttl time.Duration) *sturdyc.Client[string] {

	return sturdyc.New[string](capacity, shards, ttl, evictionPercentage,
		sturdyc.WithEarlyRefreshes(minRefreshDelay, maxRefreshDelay, retryBaseDelay),
		sturdyc.WithRefreshCoalescing(batchSize, batchBufferTimeout))
}

func SetLocalKey[T any](c *sturdyc.Client[string], ctx context.Context, key string, value *T) {
	json, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	c.Set(key, string(json))
}

func DeleteLocalKey(c *sturdyc.Client[string], ctx context.Context, key string) {
	c.Delete(key)
}
