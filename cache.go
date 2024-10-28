package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/creativecreature/sturdyc"
)

const (
	capacity = 10000
	// Number of shards to use for the cache.
	numShards = 20
	// Time-to-live for cache entries.
	ttl = 2 * time.Minute
	// Percentage of entries to evict when the cache is full.
	evictionPercentage = 10
	// Set a minimum and maximum refresh delay for the cache.
	minRefreshDelay = 15 * time.Minute
	maxRefreshDelay = 30 * time.Minute
	// The base for exponential backoff when retrying a refresh.
	retryBaseDelay = time.Second * 1

	// With refresh coalescing enabled, the cache will buffer refreshes
	// until the batch size is reached or the buffer timeout is hit.
	batchSize          = 10
	batchBufferTimeout = time.Second * 30
)

func NewCacheClient() *sturdyc.Client[string] {

	return sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithEarlyRefreshes(minRefreshDelay, maxRefreshDelay, retryBaseDelay),
		sturdyc.WithRefreshCoalescing(batchSize, batchBufferTimeout))

}

func SetLocal[T any](c *sturdyc.Client[string], ctx context.Context, key string, value *T) {
	json, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	c.Set(key, string(json))
}
