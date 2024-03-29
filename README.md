# gocache
golang loadable cache   [中文说明](README_CN.md)


## Features

* ✅ Multiple built-in caches that can be used individually or in combination
* ✅ Use generics to avoid duplicate code and be type safe.
* ✅ Drawing on the strengths of [gocache](https://github.com/eko/gocache) and [go-zero](https://github.com/zeromicro/go-zero) cache designs.
* ✅ Breakthrough proof: singleflight design ensures that only one request accesses the backend
* ✅ Avalanche-proof: cache expiration time is randomized to avoid a large number of cache failures at the same time
* ✅ Configurable: customizable cache key, cache configuration parameters can be modified
* ✅ Stable and reliable: already used in production environments

## Built-in cache

* [MemroyCache](memory_cache.go) (local memory based cache)
* [RedisCache](redis_cache.go) (github.com/redis/go-redis/v9 based cache)
* [ChainCache](chain_cache.go) (chained cache, can combine MemoryCache and RedisCache)
* [LoadableCache](loadable_cache) (auto-loadable cache)
* [LoadableL2Cache](loadable_l2_cache.go) (cache that combines LoadableCache and ChainCache)

# Installation

```
go get github.com/nzai/gocache
```

## Usage

### Use MemoryCache

```go
import (
    "context"
    "time"

    "github.com/nzai/gocache"
)

func main() {
    ctx := context.Background()
    mc := gocache.NewMemoryCache[string](10 * time.Second)
    err := mc.Set(ctx, "key1", "v1")
    if err ! = nil {
        log.Fatalf("failed to set cache value due to %v", err)
    }

    got, err := mc.Get(ctx, "key1")
    Get(ctx, "key1") if err ! = nil {
        log.Fatalf("failed to get cache value due to %v", err)
    }

    log.Printf("got cached value %s", got)
}
```

### MemoryCache caching objects

```go
type User struct {
    ID string
    Name string
}

mc := gocache.NewMemoryCache[*User](30 * time.Second)
```

### Use RedisCache

```go
import (
    "github.com/nzai/gocache"
    "github.com/redis/go-redis/v9"
)

client := redis.NewClient(&redis.Options{
    Addr: "127.0.0.1:6379",
})

rc := gocache.NewRedisCache[*User](client, 30 * time.Second)
```

### Use ChainCache

```go
client := redis.NewClient(&redis.Options{
    Addr: "127.0.0.1:6379",
})

mc := gocache.NewMemoryCache[string](10 * time.Second)
rc := gocache.NewRedisCache[string](client, 30 * time.Second)
cc := gocache.NewChainCache[string](mc, rc)
```

### Use LoadableCache

```go
type Request struct {
    ID string
}

type Response struct {
    Name string
    Value string
}

func GetValue(request *Request) (*Response, error) {
    return &Response{Name:"n1", Value:"v1"}
}

mc := gocache.NewMemoryCache[*Response](10 * time.Minute)
lc := gocache.NewLoadableCache[*Request, *Response](mc)

request := &Request{ID: "id1"}
response, err := lc.Load(GetValue, request)
if err != nil {
    log.Fatalf("failed to get value from cache due to %v", err)
}
```

### Use LoadableL2Cache

```go
type Request struct {
    ID string
}

type Response struct {
    Name string
    Value string
}

func GetValue(request *Request) (*Response, error) {
    return &Response{Name:"n1", Value:"v1"}
}

client := redis.NewClient(&redis.Options{
    Addr: "127.0.0.1:6379",
})

lc := gocache.NewLoadableL2Cache[*Request, *Response](client, 1*time.Minute)

ctx := context.Background()
request := &Request{ID: "id1"}
response, err := lc.LoadCtx(ctx, GetValue, request)
if err != nil {
    log.Fatalf("failed to get value from cache due to %v", err)
}
```