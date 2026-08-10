package gocache

import (
	"crypto/md5"
	"fmt"
	"reflect"
)

type CacheKeyGenerator interface {
	GetCacheKey() string
}

// GenerateCacheKey returns the cache key for the given key object by returning
// the key if type is string or by computing a checksum of key structure
// if its type is other than string
func GenerateCacheKey(arg any) string {
	switch v := arg.(type) {
	case CacheKeyGenerator:
		return v.GetCacheKey()
	case string:
		return v
	default:
		return checksum(arg)
	}
}

// checksum hashes a given object into a string
func checksum(arg any) string {
	digester := md5.New()
	fmt.Fprint(digester, reflect.TypeOf(arg))
	fmt.Fprint(digester, arg)
	hash := digester.Sum(nil)

	return fmt.Sprintf("%x", hash)
}
