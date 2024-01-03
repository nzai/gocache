# gocache
golang loadable cache

## 特点

* ✅ 多个内置的缓存，可以单独使用也可以组合
* ✅ 使用泛型，避免重复代码并且类型安全
* ✅ 博采众长，参考了[gocache](https://github.com/eko/gocache)及[go-zero](https://github.com/zeromicro/go-zero)缓存设计的优点
* ✅ 防击穿：singleflight设计保证只有一个请求访问后端
* ✅ 防雪崩：缓存过期时间随机化，避免同一时间大量缓存失效
* ✅ 可配置：可自定义缓存key，缓存配置参数可修改
* ✅ 稳定可靠：已在生产环境使用

## 内置的缓存

* [Memroy](memory_cache.go) (基于本地内存的缓存)
* [Redis](redis_cache.go) (基于github.com/redis/go-redis/v9的缓存)
* [ChainCache](chain_cache.go) (链式缓存，可以组合Memory和Redis)
* [LoadableCache](loadable_cache) (可自动更新的缓存)
* [LoadableL2Cache](loadable_l2_cache.go) (整合Loadable和ChainCache的缓存)

# 安装

```
go get github.com/nzai/gocache
```

## 使用
### 使用内存缓存

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
    if err != nil {
        log.Fatalf("failed to set cache value due to %v", err)
    }

    got, err := mc.Get(ctx, "key1")
    if err != nil {
        log.Fatalf("failed to get cache value due to %v", err)
    }

    log.Printf("got cached value %s", got)
}
```

### 使用内存缓存对象

```go
type User struct {
    ID string
    Name string
}

mc := gocache.NewMemoryCache[*User](30 * time.Second)
```

### 使用Redis缓存

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

### 使用ChainCache构造二级缓存

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

### 使用LoadableCache

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

mc := gocache.NewMemoryCache[*Response](10 * time.Minute)
lc := gocache.NewLoadableCache[*Request, *Response](mc)

request := &Request{ID: "id1"}
response, err := lc.Load(GetValue, request)
if err != nil {
    log.Fatalf("failed to get value from cache due to %v", err)
}
```

### 使用LoadableL2Cache

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

lc := gocache.NewLoadableL2Cache[*Request, *Response](client, 1*time.Minute)

ctx := context.Background()
request := &Request{ID: "id1"}
response, err := lc.LoadCtx(ctx, GetValue, request)
if err != nil {
    log.Fatalf("failed to get value from cache due to %v", err)
}
```