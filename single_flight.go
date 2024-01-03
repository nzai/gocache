package gocache

import (
	"context"
	"sync"
)

type SingleFlight[T, K any] interface {
	Do(fn LoadFunction[T, K], arg T) (K, error)
	DoEx(fn LoadFunction[T, K], arg T) (val K, fresh bool, err error)
	DoCtx(ctx context.Context, fn LoadFunctionCtx[T, K], arg T) (K, error)
	DoExCtx(ctx context.Context, fn LoadFunctionCtx[T, K], arg T) (val K, fresh bool, err error)
}

type singleFlightGroup[T, K any] struct {
	calls map[string]*call[K]
	lock  sync.Mutex
}

type call[T any] struct {
	wg  sync.WaitGroup
	val T
	err error
}

// NewSingleFlight returns a generic single flight.
func NewSingleFlight[T, K any]() SingleFlight[T, K] {
	return &singleFlightGroup[T, K]{
		calls: make(map[string]*call[K]),
	}
}

func (g *singleFlightGroup[T, K]) Do(fn LoadFunction[T, K], arg T) (K, error) {
	val, _, err := g.DoEx(fn, arg)
	return val, err
}

func (g *singleFlightGroup[T, K]) DoCtx(ctx context.Context, fn LoadFunctionCtx[T, K], arg T) (K, error) {
	val, _, err := g.DoExCtx(ctx, fn, arg)
	return val, err
}

func (g *singleFlightGroup[T, K]) DoEx(fn LoadFunction[T, K], arg T) (val K, fresh bool, err error) {
	key := GenerateCacheKey(arg)
	c, done := g.createCall(key)
	if done {
		return c.val, false, c.err
	}

	g.makeCall(c, key, fn, arg)
	return c.val, true, c.err
}

func (g *singleFlightGroup[T, K]) DoExCtx(ctx context.Context, fn LoadFunctionCtx[T, K], arg T) (val K, fresh bool, err error) {
	key := GenerateCacheKey(arg)
	c, done := g.createCall(key)
	if done {
		return c.val, false, c.err
	}

	g.makeCallCtx(ctx, c, key, fn, arg)
	return c.val, true, c.err
}

func (g *singleFlightGroup[T, K]) createCall(key string) (c *call[K], done bool) {
	g.lock.Lock()
	if c, ok := g.calls[key]; ok {
		g.lock.Unlock()
		c.wg.Wait()
		return c, true
	}

	c = new(call[K])
	c.wg.Add(1)
	g.calls[key] = c
	g.lock.Unlock()

	return c, false
}

func (g *singleFlightGroup[T, K]) makeCall(c *call[K], key string, fn LoadFunction[T, K], arg T) {
	defer func() {
		g.lock.Lock()
		delete(g.calls, key)
		g.lock.Unlock()
		c.wg.Done()
	}()

	c.val, c.err = fn(arg)
}

func (g *singleFlightGroup[T, K]) makeCallCtx(ctx context.Context, c *call[K], key string, fn LoadFunctionCtx[T, K], arg T) {
	defer func() {
		g.lock.Lock()
		delete(g.calls, key)
		g.lock.Unlock()
		c.wg.Done()
	}()

	c.val, c.err = fn(ctx, arg)
}
