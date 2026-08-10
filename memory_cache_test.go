package gocache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
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
		sleep: 300 * time.Millisecond,
		want:  "vvvvv",
	}, {
		name:  "key过期",
		key:   "k2",
		value: "vvvvv",
		sleep: 1500 * time.Millisecond,
		want:  "",
	}}

	ctx := context.Background()
	ms := NewMemoryCache[string](time.Second)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ms.Set(ctx, tt.key, tt.value); err != nil {
				t.Errorf("MemoryCache.Set() error = %v", err)
			}

			time.Sleep(tt.sleep)

			got, err := ms.Get(ctx, tt.key)
			if err != nil && err != ErrRecordNotFound {
				t.Errorf("MemoryCache.Get() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("MemoryCache.Get() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	ctx := context.Background()
	ms := NewMemoryCache[string](time.Second)

	if err := ms.Set(ctx, "k1", "v1"); err != nil {
		t.Errorf("MemoryCache.Set() error = %v", err)
	}

	if err := ms.Delete(ctx, "k1"); err != nil {
		t.Errorf("MemoryCache.Delete() error = %v", err)
	}

	if _, err := ms.Get(ctx, "k1"); err != ErrRecordNotFound {
		t.Errorf("MemoryCache.Get() after Delete error = %v, want = %v", err, ErrRecordNotFound)
	}
}

func TestMemoryCache_WithKeyPrefix(t *testing.T) {
	ctx := context.Background()
	ms := NewMemoryCache[string](time.Second, WithKeyPrefix("prefix:"))

	if err := ms.Set(ctx, "k1", "v1"); err != nil {
		t.Errorf("MemoryCache.Set() error = %v", err)
	}

	// the prefix is applied transparently, the unprefixed key still hits
	got, err := ms.Get(ctx, "k1")
	if err != nil {
		t.Errorf("MemoryCache.Get() error = %v", err)
	}
	if got != "v1" {
		t.Errorf("MemoryCache.Get() got = %v, want = %v", got, "v1")
	}

	// a key that already carries the prefix is treated as a different key
	if _, err := ms.Get(ctx, "prefix:k1"); err != ErrRecordNotFound {
		t.Errorf("MemoryCache.Get() with prefix error = %v, want = %v", err, ErrRecordNotFound)
	}

	if err := ms.Delete(ctx, "k1"); err != nil {
		t.Errorf("MemoryCache.Delete() error = %v", err)
	}

	if _, err := ms.Get(ctx, "k1"); err != ErrRecordNotFound {
		t.Errorf("MemoryCache.Get() after Delete error = %v, want = %v", err, ErrRecordNotFound)
	}
}

func TestNewMemoryCache_InvalidExpiration(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewMemoryCache(0) expected panic, got nil")
		}
	}()

	NewMemoryCache[string](0)
}

func BenchmarkMemoryCache_GetString(b *testing.B) {
	ctx := context.Background()
	ms := NewMemoryCache[string](1 * time.Minute)

	key := "k1234"
	value := "4321value"
	err := ms.Set(ctx, key, value)
	if err != nil {
		b.Errorf("failed to set value due to %v", err)
	}

	for index := 0; index < b.N; index++ {
		ms.Get(ctx, key)
	}
}

func BenchmarkMemoryCache_GetObject(b *testing.B) {
	ctx := context.Background()
	ms := NewMemoryCache[*getResponse](1 * time.Minute)

	key := "k222"
	err := ms.Set(ctx, key, &getResponse{})
	if err != nil {
		b.Errorf("failed to set value due to %v", err)
	}

	for index := 0; index < b.N; index++ {
		ms.Get(ctx, key)
	}
}
