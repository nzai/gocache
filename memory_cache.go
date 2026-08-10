package gocache

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/nzai/timewheel"
)

// entry holds the cached value and a generation number that increases on
// every Set of the key, so an expired timer callback can tell whether the
// entry it observed is still the current one before deleting it.
type entry[T any] struct {
	value T
	gen   uint64
}

type MemoryCache[T any] struct {
	config          *CacheConfig
	data            map[string]*entry[T]
	lock            *sync.Mutex
	timingWheel     *timewheel.TimeWheel
	expiration      time.Duration
	expiryDeviation float64
	genCounter      uint64
	stopOnce        sync.Once
}

func NewMemoryCache[T any](expiration time.Duration, options ...CacheOption) *MemoryCache[T] {
	if expiration <= 0 {
		panic("gocache: NewMemoryCache expiration must be positive")
	}

	s := &MemoryCache[T]{
		config:          &CacheConfig{},
		data:            make(map[string]*entry[T]),
		lock:            &sync.Mutex{},
		expiration:      expiration,
		expiryDeviation: ExpiryDeviation,
	}

	for _, option := range options {
		option(s.config)
	}

	// keep the timing wheel ticking at a sane positive interval, a too-small
	// interval would make the ticker busy-loop; the wheel fires entries
	// within one interval after their expiration, so a small interval keeps
	// the actual expiration close to the nominal one
	baseInterval := expiration / 10
	if baseInterval < time.Millisecond {
		baseInterval = time.Millisecond
	}

	s.timingWheel = timewheel.NewTimeWheel(baseInterval, 60, func(key string, value any) {
		gen := value.(uint64)
		s.lock.Lock()
		// only delete the entry if it is still the one this timer scheduled,
		// a newer Set may have replaced it or a Delete may have removed it
		if e, ok := s.data[key]; ok && e.gen == gen {
			delete(s.data, key)
		}
		s.lock.Unlock()
	})

	return s
}

func (s *MemoryCache[T]) Set(ctx context.Context, key string, value T) error {
	if s.config.Prefix != "" {
		key = s.config.Prefix + key
	}

	// interval [0.95, 1.05)
	deviation := 1.0 - s.expiryDeviation + rand.Float64()*s.expiryDeviation*2
	expiration := time.Duration(float64(s.expiration) * deviation)

	s.lock.Lock()
	e, found := s.data[key]
	if !found {
		e = &entry[T]{}
		s.data[key] = e
	}
	e.value = value
	e.gen = s.nextGen()
	// update the timing wheel while holding the data lock, so that Set/Delete
	// and the expiry callback can never interleave
	s.timingWheel.Set(key, e.gen, expiration)
	s.lock.Unlock()

	return nil
}

func (s *MemoryCache[T]) Get(ctx context.Context, key string) (T, error) {
	if s.config.Prefix != "" {
		key = s.config.Prefix + key
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	e, ok := s.data[key]
	if ok {
		return e.value, nil
	}

	var zero T
	return zero, ErrRecordNotFound
}

func (s *MemoryCache[T]) Delete(ctx context.Context, key string) error {
	if s.config.Prefix != "" {
		key = s.config.Prefix + key
	}

	s.lock.Lock()
	delete(s.data, key)
	s.timingWheel.Delete(key)
	s.lock.Unlock()

	return nil
}

// nextGen returns a monotonically increasing generation number so that the
// generation of a freshly created entry can never collide with a pending
// timer scheduled for the same key before it was deleted.
func (s *MemoryCache[T]) nextGen() uint64 {
	s.genCounter++
	return s.genCounter
}

// Stop stops the expiry timing wheel. Entries already in the cache stay there
// but will never expire; the cache must not be used after Stop is called.
func (s *MemoryCache[T]) Stop() {
	s.stopOnce.Do(func() {
		s.timingWheel.Stop()
	})
}
