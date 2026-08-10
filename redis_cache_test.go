package gocache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// requireRedis returns a redis client connected to the local test instance,
// or skips the test when no redis server is available at 127.0.0.1:6379.
func requireRedis(t *testing.T) *redis.Client {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not available, skip: %v", err)
	}

	return client
}

func TestRedisCache_SetTTLAndGet(t *testing.T) {
	expiration := time.Second
	tests := []struct {
		name  string
		key   string
		value string
		sleep time.Duration
		want  string
	}{{
		name:  "key没有过期",
		key:   "k1",
		value: "vvvvv",
		sleep: 0,
		want:  "vvvvv",
	}, {
		name:  "key过期",
		key:   "k2",
		value: "vvvvv",
		sleep: expiration + 1*time.Second,
		want:  "",
	}}

	ctx := context.Background()
	client := requireRedis(t)

	rs := NewRedisCache[string](client, expiration)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := rs.Set(ctx, tt.key, tt.value); err != nil {
				t.Errorf("RedisCache.Set() error = %v", err)
			}

			time.Sleep(tt.sleep)

			got, err := rs.Get(ctx, tt.key)
			if err != nil && err != ErrRecordNotFound {
				t.Errorf("RedisCache.Get() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("RedisCache.Get() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestNewRedisCache_InvalidExpiration(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewRedisCache(0) expected panic, got nil")
		}
	}()

	NewRedisCache[string](nil, 0)
}

func BenchmarkRedisCache_GetString(b *testing.B) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	rs := NewRedisCache[string](client, 1*time.Minute)

	key := "k1234"
	value := "4321value"
	err := rs.Set(ctx, key, value)
	if err != nil {
		b.Errorf("failed to set value due to %v", err)
	}

	for index := 0; index < b.N; index++ {
		rs.Get(ctx, key)
	}
}

func BenchmarkRedisCache_GetObject(b *testing.B) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	rs := NewRedisCache[*getResponse](client, 1*time.Minute)

	key := "k222"
	err := rs.Set(ctx, key, &getResponse{})
	if err != nil {
		b.Errorf("failed to set value due to %v", err)
	}

	for index := 0; index < b.N; index++ {
		rs.Get(ctx, key)
	}
}
