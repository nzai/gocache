package gocache

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache[T any] struct {
	config          *CacheConfig
	client          *redis.Client
	expiration      time.Duration
	expiryDeviation float64
}

func NewRedisCache[T any](client *redis.Client, expiration time.Duration, options ...CacheOption) *RedisCache[T] {
	if expiration <= 0 {
		panic("gocache: NewRedisCache expiration must be positive")
	}

	s := &RedisCache[T]{
		config:          &CacheConfig{},
		client:          client,
		expiration:      expiration,
		expiryDeviation: ExpiryDeviation,
	}

	for _, option := range options {
		option(s.config)
	}

	return s
}

func (s RedisCache[T]) Set(ctx context.Context, key string, value T) error {
	marshaled, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// interval [0.95, 1.05)
	deviation := 1.0 - s.expiryDeviation + rand.Float64()*s.expiryDeviation*2
	expiration := time.Duration(float64(s.expiration) * deviation)

	if s.config.Prefix != "" {
		key = s.config.Prefix + key
	}

	return s.client.Set(ctx, key, string(marshaled), expiration).Err()
}

func (s RedisCache[T]) Get(ctx context.Context, key string) (value T, err error) {
	if s.config.Prefix != "" {
		key = s.config.Prefix + key
	}

	marshaled, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return value, ErrRecordNotFound
	}
	if err != nil {
		return value, err
	}

	err = json.Unmarshal([]byte(marshaled), &value)
	if err != nil {
		return value, err
	}

	return value, nil
}

func (s RedisCache[T]) Delete(ctx context.Context, key string) error {
	if s.config.Prefix != "" {
		key = s.config.Prefix + key
	}

	return s.client.Del(ctx, key).Err()
}
