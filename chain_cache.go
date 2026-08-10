package gocache

import (
	"context"
)

type ChainCacheValue[T any] struct {
	Caches []Cache[T]
	Key    string
	Value  T
}

type ChainCache[T any] struct {
	caches       []Cache[T]
	singleFlight SingleFlight[string, T]
}

// NewChainCache instantiates a new cache that combines other caches
func NewChainCache[T any](caches ...Cache[T]) *ChainCache[T] {
	if len(caches) == 0 {
		panic("caches can't be empty")
	}

	return &ChainCache[T]{
		caches:       caches,
		singleFlight: NewSingleFlight[string, T](),
	}
}

func (c ChainCache[T]) Set(ctx context.Context, key string, value T) error {
	var err error
	for index := len(c.caches) - 1; index >= 0; index-- {
		err = c.caches[index].Set(ctx, key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c ChainCache[T]) Get(ctx context.Context, key string) (value T, err error) {
	return c.singleFlight.DoCtx(ctx, func(ctx context.Context, key string) (T, error) {
		for index, cache := range c.caches {
			value, err = cache.Get(ctx, key)
			if err == ErrRecordNotFound {
				continue
			}
			if err != nil {
				return value, err
			}

			// refresh previous caches
			for i := 0; i < index; i++ {
				err = c.caches[i].Set(ctx, key, value)
				if err != nil {
					continue
				}
			}

			return value, nil
		}

		return value, ErrRecordNotFound
	}, key)
}

func (c ChainCache[T]) Delete(ctx context.Context, key string) error {
	var err error
	for index := len(c.caches) - 1; index >= 0; index-- {
		if e := c.caches[index].Delete(ctx, key); e != nil && err == nil {
			err = e
		}
	}

	return err
}
