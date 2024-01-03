package gocache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type LoadableL2Cache[T, K any] struct {
	*LoadableCache[T, K]
}

func NewLoadableL2Cache[T, K any](client *redis.Client, expiration time.Duration) *LoadableL2Cache[T, K] {
	return &LoadableL2Cache[T, K]{
		NewLoadableCache[T, K](
			NewChainCache[K](
				NewMemoryCache[K](expiration/4),      // L1 cache
				NewRedisCache[K](client, expiration), // L2 cache
			),
		),
	}
}
