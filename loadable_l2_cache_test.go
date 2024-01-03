package gocache

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func BenchmarkLoadableL2Cache_GetObject(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	lc := NewLoadableL2Cache[*getRequest, *getResponse](client, 1*time.Minute)
	request := &getRequest{ID: "K3"}

	s := &score{total: 1}
	for index := 0; index < b.N; index++ {
		lc.Load(s.Get, request)
	}
}
