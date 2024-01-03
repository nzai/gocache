package gocache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestChainCache_SetAndGet(t *testing.T) {
	msExpiration := time.Second
	redisExpiration := msExpiration * 10

	tests := []struct {
		name        string
		key         string
		memoryValue string
		redisValue  string
		sleep       time.Duration
		want        string
		err         error
	}{{
		name:        "memory hit",
		key:         "k1",
		memoryValue: "11",
		redisValue:  "22",
		sleep:       msExpiration / 2,
		want:        "11",
		err:         nil,
	}, {
		name:        "memory miss, redis hit",
		key:         "k2",
		memoryValue: "11",
		redisValue:  "22",
		sleep:       msExpiration + time.Second,
		want:        "22",
		err:         nil,
	}, {
		name:        "memory miss, redis hit",
		key:         "k3",
		memoryValue: "11",
		redisValue:  "22",
		sleep:       redisExpiration + time.Second,
		want:        "",
		err:         ErrRecordNotFound,
	}}

	ms := NewMemoryCache[string](msExpiration)
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	rs := NewRedisCache[string](client, redisExpiration)
	cc := NewChainCache[string](ms, rs)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ms.Set(ctx, tt.key, tt.memoryValue); err != nil {
				t.Errorf("ChainCache.Set() error = %v", err)
			}

			if err := rs.Set(ctx, tt.key, tt.redisValue); err != nil {
				t.Errorf("ChainCache.Set() error = %v", err)
			}

			time.Sleep(tt.sleep)

			got, err := cc.Get(ctx, tt.key)
			if tt.err != nil && err != tt.err {
				t.Errorf("ChainCache.Get() error got = %v, want = %v", err, tt.err)
			}

			if got != tt.want {
				t.Errorf("ChainCache.Get() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func BenchmarkChainCache_GetString(b *testing.B) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	rs := NewRedisCache[string](client, 1*time.Minute)
	ms := NewMemoryCache[string](30 * time.Second)
	cc := NewChainCache[string](ms, rs)

	key := "k1234"
	value := "4321value"
	err := cc.Set(ctx, key, value)
	if err != nil {
		b.Errorf("failed to set value due to %v", err)
	}

	for index := 0; index < b.N; index++ {
		cc.Get(ctx, key)
	}
}

func BenchmarkChainCache_GetObject(b *testing.B) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	rs := NewRedisCache[*getResponse](client, 1*time.Minute)
	ms := NewMemoryCache[*getResponse](30 * time.Second)
	cc := NewChainCache[*getResponse](ms, rs)

	key := "k222"
	err := cc.Set(ctx, key, &getResponse{})
	if err != nil {
		b.Errorf("failed to set value due to %v", err)
	}

	for index := 0; index < b.N; index++ {
		cc.Get(ctx, key)
	}
}
