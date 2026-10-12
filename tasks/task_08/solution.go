package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	l := &Limiter{
		clock: clock,
		rate:  ratePerSec,
		burst: burst,
	}
	if clock != nil {
		l.last = clock.Now()
	}
	if burst > 0 {
		l.tokens = float64(burst)
	}
	return l
}
func (l *Limiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.clock == nil || l.burst <= 0 {
		return false
	} else if n > l.burst || n <= 0 {
		return false
	} else {
		currentTime := l.clock.Now()
		if l.rate > 0 && currentTime.After(l.last) {
			diff := currentTime.Sub(l.last).Seconds()
			l.tokens += diff * l.rate
			if l.tokens > float64(l.burst) {
				l.tokens = float64(l.burst)
			}
			l.last = currentTime
		}

		if l.tokens >= float64(n) {
			l.tokens -= float64(n)
			return true
		}
		return false
	}
}
