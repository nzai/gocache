package gocache

import (
	"context"
)

var (
	// it can be changed and default value is 0.25
	DecayFactor = 0.25
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

// NewChainCache instanciates a new cache that combine other caches
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
