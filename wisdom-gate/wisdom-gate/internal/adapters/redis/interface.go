package redis

import (
	"context"
	"time"
)

type ClientInterface interface {
	StoreChallenge(ctx context.Context, token string, challenge string, ttl time.Duration) error
	GetChallenge(ctx context.Context, token string) (string, error)
	MarkChallengeSpent(ctx context.Context, token string, ttl time.Duration) (bool, error)
	DeleteChallenge(ctx context.Context, token string) error
	Close() error
}
