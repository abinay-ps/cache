package cache

import (
	"encoding/json"
	"time"

	"github.com/creativecreature/sturdyc"
)

// This function will return a new Local Cache Client.
func NewCacheClient(capacity int, shards int, batchSize int, batchBufferTimeout time.Duration,
	evictionPercentage int, maxRefreshDelay time.Duration, minRefreshDelay time.Duration,
	retryBaseDelay time.Duration, ttl time.Duration) *sturdyc.Client[string] {

	return sturdyc.New[string](capacity, shards, ttl, evictionPercentage,
		sturdyc.WithEarlyRefreshes(minRefreshDelay, maxRefreshDelay, retryBaseDelay),
		sturdyc.WithRefreshCoalescing(batchSize, batchBufferTimeout))
}

// This function will set a key-value pair in Local Cache.
func SetLocalKey[T any](c *sturdyc.Client[string], key string, value *T) {
	json, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	c.Set(key, string(json))
}

// This function will delete a key-value pair in Local Cache.
func DeleteLocalKey(c *sturdyc.Client[string], key string) {
	c.Delete(key)
}
