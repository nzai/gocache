package gocache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type addRequest struct {
	Value int64
}

type addResponse struct {
	Total int64
}

type getRequest struct {
	ID string
}

type getResponse struct {
	Value int64
}

type score struct {
	total int64
}

func (s *score) Add(request *addRequest) (*addResponse, error) {
	if request.Value < 0 {
		panic("value cannot be negative numbers")
	}

	time.Sleep(time.Millisecond * 100)
	newValue := atomic.AddInt64(&s.total, request.Value)
	return &addResponse{Total: newValue}, nil
}

func (s *score) AddCtx(ctx context.Context, request *addRequest) (*addResponse, error) {
	return s.Add(request)
}

func (s score) Get(request *getRequest) (*getResponse, error) {
	return &getResponse{Value: s.total}, nil
}

func TestSingleFlight_Do(t *testing.T) {
	tests := []struct {
		name       string
		concurrent int
		arg        int64
		want       int64
		wantErr    bool
	}{{
		name:       "1000并发",
		concurrent: 1000,
		arg:        1,
		want:       1,
		wantErr:    false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := new(score)
			sf := NewSingleFlight[*addRequest, *addResponse]()

			wg := &sync.WaitGroup{}

			for index := 0; index < tt.concurrent; index++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					response, err := sf.Do(s.Add, &addRequest{Value: tt.arg})
					if (err != nil) != tt.wantErr {
						t.Errorf("SingleFlight.Do() error = %v, wantErr %v", err, tt.wantErr)
						return
					}

					if response.Total != tt.want {
						t.Errorf("SingleFlight.Do() = %v, want %v", response.Total, tt.want)
					}
				}()
			}

			wg.Wait()
		})
	}
}

func TestSingleFlight_DoEx(t *testing.T) {
	tests := []struct {
		name       string
		concurrent int
		arg        int64
		want       int64
		wantErr    bool
	}{{
		name:       "1000并发",
		concurrent: 1000,
		arg:        1,
		want:       1,
		wantErr:    false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := new(score)
			sf := NewSingleFlight[*addRequest, *addResponse]()

			wg := &sync.WaitGroup{}

			var refreshCount int64
			for index := 0; index < tt.concurrent; index++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					response, refresh, err := sf.DoEx(s.Add, &addRequest{Value: tt.arg})
					if (err != nil) != tt.wantErr {
						t.Errorf("SingleFlight.DoEx() error = %v, wantErr %v", err, tt.wantErr)
						return
					}

					if response.Total != tt.want {
						t.Errorf("SingleFlight.DoEx() = %v, want %v", response.Total, tt.want)
					}

					if refresh {
						atomic.AddInt64(&refreshCount, 1)
					}
				}()
			}

			wg.Wait()

			if refreshCount != 1 {
				t.Errorf("SingleFlight.DoEx() refreshed more than once, freshed: %d", refreshCount)
			}
		})
	}
}

func TestSingleFlight_DoCtx(t *testing.T) {
	tests := []struct {
		name       string
		concurrent int
		arg        int64
		want       int64
		wantErr    bool
	}{{
		name:       "1000并发ctx",
		concurrent: 1000,
		arg:        1,
		want:       1,
		wantErr:    false,
	}}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := new(score)
			sf := NewSingleFlight[*addRequest, *addResponse]()

			wg := &sync.WaitGroup{}

			for index := 0; index < tt.concurrent; index++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					response, err := sf.DoCtx(ctx, s.AddCtx, &addRequest{Value: tt.arg})
					if (err != nil) != tt.wantErr {
						t.Errorf("SingleFlight.DoCtx() error = %v, wantErr %v", err, tt.wantErr)
						return
					}

					if response.Total != tt.want {
						t.Errorf("SingleFlight.DoCtx() = %v, want %v", response.Total, tt.want)
					}
				}()
			}

			wg.Wait()
		})
	}
}

func TestSingleFlight_DoExCtx(t *testing.T) {
	tests := []struct {
		name       string
		concurrent int
		arg        int64
		want       int64
		wantErr    bool
	}{{
		name:       "1000并发ctx",
		concurrent: 1000,
		arg:        1,
		want:       1,
		wantErr:    false,
	}}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := new(score)
			sf := NewSingleFlight[*addRequest, *addResponse]()

			wg := &sync.WaitGroup{}

			var refreshCount int64
			for index := 0; index < tt.concurrent; index++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					response, refresh, err := sf.DoExCtx(ctx, s.AddCtx, &addRequest{Value: tt.arg})
					if (err != nil) != tt.wantErr {
						t.Errorf("SingleFlight.DoExCtx() error = %v, wantErr %v", err, tt.wantErr)
						return
					}

					if response.Total != tt.want {
						t.Errorf("SingleFlight.DoExCtx() = %v, want %v", response.Total, tt.want)
					}

					if refresh {
						atomic.AddInt64(&refreshCount, 1)
					}
				}()
			}

			wg.Wait()

			if refreshCount != 1 {
				t.Errorf("SingleFlight.DoExCtx() refreshed more than once, freshed: %d", refreshCount)
			}
		})
	}
}
