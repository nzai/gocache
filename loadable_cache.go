package gocache

import (
	"context"
	"fmt"
)

type LoadFunction[T, K any] func(T) (K, error)
type LoadFunctionCtx[T, K any] func(context.Context, T) (K, error)

type LoadableCache[T, K any] struct {
	cache        Cache[K]
	singleFlight SingleFlight[T, K]
}

// NewLoadableCache instanciates a new cache that uses a function to load data
func NewLoadableCache[T, K any](cache Cache[K]) *LoadableCache[T, K] {
	return &LoadableCache[T, K]{
		cache:        cache,
		singleFlight: NewSingleFlight[T, K](),
	}
}

// Load returns the object stored in cache
func (c *LoadableCache[T, K]) Load(fn LoadFunction[T, K], arg T) (K, error) {
	ctx := context.Background()
	key := GenerateCacheKey(arg)
	value, err := c.cache.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	return c.singleFlight.Do(func(arg T) (value K, err error) {
		defer func() {
			if err1 := recover(); err1 != nil {
				err = fmt.Errorf("load function panic: %v", err1)
			}
		}()

		// because load function is an IO query ,which is slower than cache get
		// so we do double check, cache might be taken by another call
		value, err = c.cache.Get(ctx, key)
		if err == nil {
			return value, nil
		}

		// Unable to find in cache, try to load it from load function
		object, err := fn(arg)
		if err != nil {
			// TODO: we may need to cache errors to reduce the access to the backend
			return object, err
		}

		// Then, put it back in cache
		c.cache.Set(ctx, key, object)

		return object, nil
	}, arg)
}

// LoadCtx returns the object stored in cache with context
func (c *LoadableCache[T, K]) LoadCtx(ctx context.Context, fn LoadFunctionCtx[T, K], arg T) (K, error) {
	key := GenerateCacheKey(arg)
	value, err := c.cache.Get(ctx, key)
	if err == nil {
		return value, err
	}

	return c.singleFlight.DoCtx(ctx, func(ctx context.Context, arg T) (value K, err error) {
		defer func() {
			if err1 := recover(); err1 != nil {
				err = fmt.Errorf("load function panic: %v", err1)
			}
		}()

		// because load function is an IO query ,which is slower than cache get
		// so we do double check, cache might be taken by another call
		value, err = c.cache.Get(ctx, key)
		if err == nil {
			return value, nil
		}

		// Unable to find in cache, try to load it from load function
		object, err := fn(ctx, arg)
		if err != nil {
			// TODO: we may need to cache errors to reduce the access to the backend
			return object, err
		}

		// Then, put it back in cache
		c.cache.Set(ctx, key, object)

		return object, nil
	}, arg)
}
