package gocache

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type appender struct {
	slice []string
}

func (s *appender) Append(value string) {
	s.slice = append(s.slice, value)
}

func (s appender) Get(key string) (string, error) {
	return strings.Join(s.slice, ""), nil
}

func TestLoadableCache_LoadFromMultiCaches_string(t *testing.T) {
	expiration := 2 * time.Second

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	tests := []struct {
		name  string
		cache Cache[string]
		id    string
	}{{
		name:  "L1-memory",
		cache: NewMemoryCache[string](expiration),
	}, {
		name:  "L1-redis",
		cache: NewRedisCache[string](redisClient, expiration, WithKeyPrefix("string1:")),
	}, {
		name:  "L1-memory L2-redis",
		cache: NewChainCache[string](NewMemoryCache[string](expiration/2), NewRedisCache[string](redisClient, expiration, WithKeyPrefix("string2:"))),
	}, {
		name:  "L1-memory L2-memory",
		cache: NewChainCache[string](NewMemoryCache[string](expiration/2), NewMemoryCache[string](expiration)),
	}, {
		name:  "L1-redis L2-memory",
		cache: NewChainCache[string](NewRedisCache[string](redisClient, expiration/2, WithKeyPrefix("string3:")), NewMemoryCache[string](expiration)),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := appender{}

			lc := NewLoadableCache[string, string](tt.cache)

			response, err := lc.Load(a.Get, tt.name)
			if err != nil {
				t.Errorf("LoadableCache.Load() error = %v", err)
			}

			if response != "" {
				t.Errorf("LoadableCache.Load() got = %v, want = %v", response, "")
			}

			a.Append("updated")

			response, err = lc.Load(a.Get, tt.name)
			if err != nil {
				t.Errorf("LoadableCache.Load() error = %v", err)
			}

			if response != "" {
				t.Errorf("LoadableCache.Load() got = %v, want = %v", response, "")
			}

			time.Sleep(time.Duration(float64(expiration) * float64(1+ExpiryDeviation)))

			response, err = lc.Load(a.Get, tt.name)
			if err != nil {
				t.Errorf("LoadableCache.Load() error = %v", err)
			}

			if response != "updated" {
				t.Errorf("LoadableCache.Load() got = %v, want = %v", response, "updated")
			}
		})
	}
}

func TestLoadableCache_LoadFromMultiCaches_StructPointer(t *testing.T) {
	expiration := 2 * time.Second

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	tests := []struct {
		name  string
		cache Cache[*getResponse]
		id    string
	}{{
		name:  "L1-memory",
		cache: NewMemoryCache[*getResponse](expiration),
	}, {
		name:  "L1-redis",
		cache: NewRedisCache[*getResponse](redisClient, expiration, WithKeyPrefix("StructPointer1:")),
	}, {
		name:  "L1-memory L2-redis",
		cache: NewChainCache[*getResponse](NewMemoryCache[*getResponse](expiration/2), NewRedisCache[*getResponse](redisClient, expiration, WithKeyPrefix("StructPointer2:"))),
	}, {
		name:  "L1-memory L2-memory",
		cache: NewChainCache[*getResponse](NewMemoryCache[*getResponse](expiration/2), NewMemoryCache[*getResponse](expiration)),
	}, {
		name:  "L1-redis L2-memory",
		cache: NewChainCache[*getResponse](NewRedisCache[*getResponse](redisClient, expiration/2, WithKeyPrefix("StructPointer3:")), NewMemoryCache[*getResponse](expiration)),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &score{total: 1}

			lc := NewLoadableCache[*getRequest, *getResponse](tt.cache)

			response, err := lc.Load(s.Get, &getRequest{ID: tt.name})
			if err != nil {
				t.Errorf("LoadableCache.Load() error = %v", err)
			}

			if response.Value != 1 {
				t.Errorf("LoadableCache.Load() got = %v, want = %v", response.Value, 1)
			}

			s.Add(&addRequest{Value: 2})

			response, err = lc.Load(s.Get, &getRequest{ID: tt.name})
			if err != nil {
				t.Errorf("LoadableCache.Load() error = %v", err)
			}

			if response.Value != 1 {
				t.Errorf("LoadableCache.Load() got = %v, want = %v", response.Value, 1)
			}

			time.Sleep(time.Duration(float64(expiration) * float64(1+ExpiryDeviation)))

			response, err = lc.Load(s.Get, &getRequest{ID: tt.name})
			if err != nil {
				t.Errorf("LoadableCache.Load() error = %v", err)
			}

			if response.Value != 3 {
				t.Errorf("LoadableCache.Load() got = %v, want = %v", response.Value, 3)
			}
		})
	}
}

func TestLoadableCache_LoadFromMultiCaches_InParallel(t *testing.T) {
	expiration := 5 * time.Second

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	tests := []struct {
		name  string
		cache Cache[*getResponse]
		id    string
	}{{
		name:  "L1-memory",
		cache: NewMemoryCache[*getResponse](expiration),
	}, {
		name:  "L1-memory L2-redis",
		cache: NewChainCache[*getResponse](NewMemoryCache[*getResponse](expiration/2), NewRedisCache[*getResponse](redisClient, expiration)),
	}}

	parallelRun := func(parallel int, fn func()) {
		wg := &sync.WaitGroup{}
		wg.Add(parallel)

		for index := 0; index < parallel; index++ {
			go func() {
				defer wg.Done()
				fn()
			}()
		}

		wg.Wait()
	}

	parallel := 1000
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &score{total: 1}

			lc := NewLoadableCache[*getRequest, *getResponse](tt.cache)

			parallelRun(parallel, func() {
				response, err := lc.Load(s.Get, &getRequest{ID: tt.name})
				if err != nil {
					t.Fatalf("LoadableCache.Load() error = %v", err)
				}

				if response.Value != 1 {
					t.Fatalf("LoadableCache.Load() got = %v, want = %v", response.Value, 1)
				}
			})

			s.Add(&addRequest{Value: 2})

			parallelRun(parallel, func() {
				response, err := lc.Load(s.Get, &getRequest{ID: tt.name})
				if err != nil {
					t.Fatalf("LoadableCache.Load() error = %v", err)
				}

				if response.Value != 1 {
					t.Fatalf("LoadableCache.Load() got = %v, want = %v", response.Value, 1)
				}
			})

			time.Sleep(time.Duration(float64(expiration) * float64(1+ExpiryDeviation)))

			parallelRun(parallel, func() {
				response, err := lc.Load(s.Get, &getRequest{ID: tt.name})
				if err != nil {
					t.Fatalf("LoadableCache.Load() error = %v", err)
				}

				if response.Value != s.total {
					t.Fatalf("LoadableCache.Load() got = %v, want = %v", response.Value, s.total)
				}
			})
		})
	}
}

func TestLoadableCache_LoadFromMultiCaches_InParallelCount(t *testing.T) {
	expiration := 4 * time.Second

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	tests := []struct {
		name  string
		cache Cache[*addResponse]
		id    string
	}{{
		name:  "L1-memory L2-redis",
		cache: NewChainCache[*addResponse](NewMemoryCache[*addResponse](expiration/2), NewRedisCache[*addResponse](redisClient, expiration)),
	}}

	times := 6
	parallel := 1000
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &score{total: 0}

			lc := NewLoadableCache[*addRequest, *addResponse](tt.cache)

			ch := make(chan struct{}, parallel)
			after := time.After(time.Duration(float64(times)*float64(expiration)*(1+ExpiryDeviation)) + time.Second)
		Loop:
			for {
				select {
				case <-after:
					break Loop
				case ch <- struct{}{}:
					go func() {
						_, err := lc.Load(s.Add, &addRequest{Value: 1})
						if err != nil {
							t.Errorf("LoadableCache.Load() error = %v", err)
						}

						<-ch
					}()
				}
			}

			if s.total != int64(times) {
				t.Fatalf("LoadableCache.Load() got = %v, want = %v", s.total, times)
			}
		})
	}
}

func TestLoadableCache_Load_Panic(t *testing.T) {
	expiration := 4 * time.Second

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	tests := []struct {
		name  string
		cache Cache[*addResponse]
		id    string
	}{{
		name:  "L1-memory L2-redis",
		cache: NewChainCache[*addResponse](NewMemoryCache[*addResponse](expiration/2), NewRedisCache[*addResponse](redisClient, expiration)),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &score{total: 0}

			lc := NewLoadableCache[*addRequest, *addResponse](tt.cache)

			response, err := lc.Load(s.Add, &addRequest{Value: 1})
			if err != nil {
				t.Errorf("LoadableCache.Load() error = %v", err)
			}

			if response.Total != 1 {
				t.Errorf("LoadableCache.Load() got = %v, want = %v", response.Total, 1)
			}

			_, err = lc.Load(s.Add, &addRequest{Value: -1})
			if err == nil {
				t.Errorf("LoadableCache.Load() error = %v", err)
			}
		})
	}
}

func BenchmarkLoadableCache_GetObject(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	rs := NewRedisCache[*getResponse](client, 1*time.Minute)
	ms := NewMemoryCache[*getResponse](30 * time.Second)
	cc := NewChainCache[*getResponse](ms, rs)
	lc := NewLoadableCache[*getRequest, *getResponse](cc)

	request := &getRequest{ID: "K1"}

	s := &score{total: 1}
	for index := 0; index < b.N; index++ {
		lc.Load(s.Get, request)
	}
}
