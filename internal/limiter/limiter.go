package limiter

import (
	"context"
	"sync"
)

type Limiter struct {
	sem chan struct{}
	mu  sync.Mutex
	cur int
}

func New(max int) *Limiter {
	return &Limiter{
		sem: make(chan struct{}, max),
	}
}

func (l *Limiter) Acquire(ctx context.Context) error {
	select {
	case l.sem <- struct{}{}:
		l.mu.Lock()
		l.cur++
		l.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return context.DeadlineExceeded
	}
}

func (l *Limiter) Release() {
	l.mu.Lock()
	l.cur--
	l.mu.Unlock()
	<-l.sem
}

func (l *Limiter) Active() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cur
}
