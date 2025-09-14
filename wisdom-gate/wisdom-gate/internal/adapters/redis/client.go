package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(addr string) (ClientInterface, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) StoreChallenge(ctx context.Context, token string, challenge string, ttl time.Duration) error {
	key := fmt.Sprintf("challenge:%s", token)
	return c.rdb.Set(ctx, key, challenge, ttl).Err()
}

func (c *Client) GetChallenge(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("challenge:%s", token)
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) MarkChallengeSpent(ctx context.Context, token string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("challenge:spent:%s", token)
	return c.rdb.SetNX(ctx, key, "1", ttl).Result()
}

func (c *Client) DeleteChallenge(ctx context.Context, token string) error {
	key := fmt.Sprintf("challenge:%s", token)
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

type MockRedisClient struct {
	challenges map[string]string
	spent      map[string]bool
	closeErr   error
}

func NewMockRedisClient() ClientInterface {
	return &MockRedisClient{
		challenges: make(map[string]string),
		spent:      make(map[string]bool),
	}
}

func (m *MockRedisClient) StoreChallenge(ctx context.Context, token string, challenge string, ttl time.Duration) error {
	m.challenges[token] = challenge
	return nil
}

func (m *MockRedisClient) GetChallenge(ctx context.Context, token string) (string, error) {
	challenge, exists := m.challenges[token]
	if !exists {
		return "", context.DeadlineExceeded
	}
	return challenge, nil
}

func (m *MockRedisClient) MarkChallengeSpent(ctx context.Context, token string, ttl time.Duration) (bool, error) {
	if m.spent[token] {
		return false, nil // Уже потрачен
	}
	m.spent[token] = true
	return true, nil
}

func (m *MockRedisClient) DeleteChallenge(ctx context.Context, token string) error {
	delete(m.challenges, token)
	return nil
}

func (m *MockRedisClient) Close() error {
	return m.closeErr
}
