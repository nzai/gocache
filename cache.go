package gocache

import (
	"context"
	"errors"
)

var (
	// ExpiryDeviation controls the randomization range of the cache expiration
	// time, default value is 0.05. It must be set before creating any cache
	// instance; changing it at runtime has no effect on existing caches.
	ExpiryDeviation = 0.05

	ErrRecordNotFound = errors.New("record not found")
)

type Cache[T any] interface {
	Set(ctx context.Context, key string, value T) error
	Get(ctx context.Context, key string) (T, error)
	Delete(ctx context.Context, key string) error
}
