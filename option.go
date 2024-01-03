package gocache

type CacheConfig struct {
	// key prefix
	Prefix string
}

type CacheOption func(*CacheConfig)

func WithKeyPrefix(prefix string) CacheOption {
	return func(sc *CacheConfig) {
		sc.Prefix = prefix
	}
}
