# gocache
golang loadable cache   [中文说明](README_CN.md)


## Features

* ✅ Multiple built-in caches that can be used individually or in combination
* ✅ Use of generic types to avoid duplicate code and type-safe.
* ✅ Draws on the best of [gocache](https://github.com/eko/gocache) and [go-zero](https://github.com/zeromicro/go-zero) cache designs.
* ✅ Has been used in production environments, stable and reliable

## Built-in cache

* [Memroy](memory_cache.go) (local memory based cache)
* [Redis](redis_cache.go) (github.com/redis/go-redis/v9 based cache)
* [ChainCache](chain_cache.go) (chained cache, can combine Memory and Redis)
* [LoadableCache)](loadable_cache) (auto-updatable cache)
* [LoadableL2Cache](loadable_l2_cache.go) (cache that combines Loadable and ChainCache)

# Installation

```
go get github.com/nzai/gocache
```

## Usage

### Use in-memory cache

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

### in-memory caching objects

```go
type User struct {
    ID string
    Name string
}

mc := gocache.NewMemoryCache[*User](30 * time.Second)
```

### Use Redis cache

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
import (
    "github.com/nzai/gocache"
    "github.com/redis/go-redis/v9"
)

client := redis.NewClient(&redis.Options{
    Addr: "127.0.0.1:6379",
})

mc := gocache.NewMemoryCache[string](10 * time.Second)
rc := gocache.NewRedisCache[string](client, 30 * time.Second)
cc := gocache.NewChainCache[string](mc, rc)
```

### Use LoadableCache

```go
import (
    "github.com/nzai/gocache"
)

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

mc := gocache.NewMemoryCache[*Response*](10 * time.Minute)
lc := gocache.NewLoadableCache[*Request*, *Response](mc)

request := &Request{ID: "id1"}
response, err := lc.Load(GetValue, request)
if err != nil {
    log.Fatalf("failed to get value from cache due to %v", err)
}
```

### Use LoadableL2Cache

```go
import (
    "github.com/nzai/gocache"
)

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

lc := gocache.NewLoadableL2Cache[*Request*, *Response](client, 1*time.Minute)

ctx := context.Background()
request := &Request{ID: "id1"}
response, err := lc.LoadCtx(ctx, GetValue, request)
if err != nil {
    log.Fatalf("failed to get value from cache due to %v", err)
}
```