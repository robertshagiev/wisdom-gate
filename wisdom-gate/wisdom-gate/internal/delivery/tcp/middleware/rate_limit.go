package middleware

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	protocolUC "wisdom-gate/internal/application/protocol/usecase"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) IsAllowed(clientAddr string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	if requests, exists := rl.requests[clientAddr]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[clientAddr] = validRequests
	}

	requests := rl.requests[clientAddr]
	if len(requests) >= rl.limit {
		return false
	}

	rl.requests[clientAddr] = append(requests, now)
	return true
}

func RateLimitMiddleware(limiter *RateLimiter) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error {
			if !limiter.IsAllowed(clientAddr) {
				return fmt.Errorf("rate limit exceeded")
			}
			return next(ctx, conn, clientAddr, msg)
		}
	}
}
