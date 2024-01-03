package gocache

import (
	"context"
	"errors"
)

var (
	// it can be changed and default value is 0.05
	ExpiryDeviation = 0.05

	ErrRecordNotFound = errors.New("record not found")
)

type Cache[T any] interface {
	Set(ctx context.Context, key string, value T) error
	Get(ctx context.Context, key string) (T, error)
}
