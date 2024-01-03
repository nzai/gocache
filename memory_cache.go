package gocache

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/collection"
)

type MemoryCache[T any] struct {
	config      *CacheConfig
	data        map[string]T
	lock        *sync.Mutex
	timingWheel *collection.TimingWheel
	expiration  time.Duration
}

func NewMemoryCache[T any](expiration time.Duration, options ...CacheOption) *MemoryCache[T] {
	s := &MemoryCache[T]{
		config:     &CacheConfig{},
		data:       make(map[string]T),
		lock:       &sync.Mutex{},
		expiration: expiration,
	}

	for _, option := range options {
		option(s.config)
	}

	s.timingWheel, _ = collection.NewTimingWheel(time.Second, 500, func(k, v any) {
		key, ok := k.(string)
		if !ok {
			return
		}

		s.Delete(key)
	})

	return s
}

func (s *MemoryCache[T]) Set(ctx context.Context, key string, value T) error {
	// interval [0.95, 1.05)
	deviation := 1.0 - ExpiryDeviation + rand.Float64()*ExpiryDeviation*2
	expiration := time.Duration(float64(s.expiration) * deviation)

	s.lock.Lock()
	_, found := s.data[key]
	s.data[key] = value
	s.lock.Unlock()

	if found {
		return s.timingWheel.MoveTimer(key, expiration)
	}

	return s.timingWheel.SetTimer(key, value, expiration)
}

func (s MemoryCache[T]) Get(ctx context.Context, key string) (T, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	value, ok := s.data[key]
	if ok {
		return value, nil
	}

	return value, ErrRecordNotFound
}

func (s MemoryCache[T]) Delete(key string) error {
	s.lock.Lock()
	delete(s.data, key)
	s.lock.Unlock()
	s.timingWheel.RemoveTimer(key)

	return nil
}
