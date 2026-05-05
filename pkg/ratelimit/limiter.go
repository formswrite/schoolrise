package ratelimit

import (
	"context"
	"sync"
	"time"

	"encore.dev/beta/errs"
	"golang.org/x/time/rate"
)

type bucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rps     rate.Limit
	burst   int
	ttl     time.Duration
}

func NewLimiter(perMinute int, burst int) *Limiter {
	rps := rate.Limit(float64(perMinute) / 60.0)
	if burst < 1 {
		burst = 1
	}
	l := &Limiter{
		buckets: make(map[string]*bucket),
		rps:     rps,
		burst:   burst,
		ttl:     10 * time.Minute,
	}
	go l.evictLoop()
	return l
}

func (l *Limiter) Allow(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	l.mu.Lock()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{limiter: rate.NewLimiter(l.rps, l.burst)}
		l.buckets[key] = b
	}
	b.lastSeen = time.Now()
	l.mu.Unlock()

	if !b.limiter.Allow() {
		return &errs.Error{
			Code:    errs.ResourceExhausted,
			Message: "too many requests, please slow down",
		}
	}
	return nil
}

func (l *Limiter) evictLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		l.mu.Lock()
		for k, b := range l.buckets {
			if now.Sub(b.lastSeen) > l.ttl {
				delete(l.buckets, k)
			}
		}
		l.mu.Unlock()
	}
}
