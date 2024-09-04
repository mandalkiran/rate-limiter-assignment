package main

import (
	"math"
	"sync"
	"time"
)

// Limit is represented as number of events per minute.
type Limit float64

// Inf is the infinite rate limit; it allows all events (even if burst is zero).
const Inf = Limit(math.MaxFloat64)

// A Limiter controls how frequently events are allowed to happen.
type Limiter struct {
	mu     sync.Mutex
	limit  Limit
	burst  int
	tokens float64
	// last is the last time the limiter's tokens field was updated
	last time.Time
}

// NewLimiter returns a new Limiter that allows events up to rate r and permits
// bursts of at most b tokens.
func NewLimiter(r Limit, b int) *Limiter {
	return &Limiter{
		limit: r,
		burst: b,
	}
}

// Allow reports whether an event may happen now.
func (lim *Limiter) Allow() bool {
	n := 1
	t := time.Now()

	lim.mu.Lock()
	defer lim.mu.Unlock()
	if lim.limit == Inf {
		return true
	} else if lim.limit == 0 {
		var ok bool
		if lim.burst >= n {
			ok = true
			lim.burst -= n
		}
		return ok
	}

	t, tokens := lim.advance(t)

	// Calculate the remaining number of tokens resulting from the request.
	tokens -= float64(n)

	// Decide result
	ok := tokens > 0

	if ok {
		lim.last = t
		lim.tokens = tokens
	}

	return ok
}

// advance calculates and returns an updated state for
// lim resulting from the passage of time.
func (lim *Limiter) advance(t time.Time) (newT time.Time, newTokens float64) {
	last := lim.last

	// Calculate the new number of tokens, due to time that passed.
	elapsed := t.Sub(last)
	delta := lim.limit.tokensFromDuration(elapsed)
	tokens := lim.tokens + delta
	if burst := float64(lim.burst); tokens > burst {
		tokens = burst
	}
	return t, tokens
}

// tokensFromDuration is a unit conversion function from a time duration to the number of tokens
// which could be accumulated during that duration at a rate of limit tokens per minute.
func (limit Limit) tokensFromDuration(d time.Duration) float64 {
	if limit <= 0 {
		return 0
	}
	return d.Minutes() * float64(limit)
}

// SetLimit sets a new Limit for the limiter.
func (lim *Limiter) SetLimit(newLimit Limit) {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	t, tokens := lim.advance(time.Now())

	lim.last = t
	lim.tokens = tokens
	lim.limit = newLimit
}

// SetBurst SetBurstAt sets a new burst size for the limiter.
func (lim *Limiter) SetBurst(newBurst int) {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	t, tokens := lim.advance(time.Now())

	lim.last = t
	lim.tokens = tokens
	lim.burst = newBurst
}
