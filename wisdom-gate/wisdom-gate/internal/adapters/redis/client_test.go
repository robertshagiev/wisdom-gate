package redis

import (
	"context"
	"testing"
	"time"
)

func TestMockRedisClient_StoreChallenge(t *testing.T) {
	client := NewMockRedisClient()
	ctx := context.Background()

	err := client.StoreChallenge(ctx, "test-token", "test-challenge", time.Minute)
	if err != nil {
		t.Errorf("StoreChallenge() error = %v", err)
	}

	// Проверяем что challenge сохранился
	challenge, err := client.GetChallenge(ctx, "test-token")
	if err != nil {
		t.Errorf("GetChallenge() error = %v", err)
	}

	if challenge != "test-challenge" {
		t.Errorf("GetChallenge() = %v, want %v", challenge, "test-challenge")
	}
}

func TestMockRedisClient_MarkChallengeSpent(t *testing.T) {
	client := NewMockRedisClient()
	ctx := context.Background()

	// Первый раз должен вернуть true
	spent, err := client.MarkChallengeSpent(ctx, "test-token", time.Minute)
	if err != nil {
		t.Errorf("MarkChallengeSpent() error = %v", err)
	}
	if !spent {
		t.Error("MarkChallengeSpent() = false, want true")
	}

	// Второй раз должен вернуть false (уже потрачен)
	spent, err = client.MarkChallengeSpent(ctx, "test-token", time.Minute)
	if err != nil {
		t.Errorf("MarkChallengeSpent() error = %v", err)
	}
	if spent {
		t.Error("MarkChallengeSpent() = true, want false")
	}
}

func TestMockRedisClient_GetChallenge_NotFound(t *testing.T) {
	client := NewMockRedisClient()
	ctx := context.Background()

	_, err := client.GetChallenge(ctx, "non-existent-token")
	if err == nil {
		t.Error("GetChallenge() error = nil, want error")
	}
}

func TestMockRedisClient_DeleteChallenge(t *testing.T) {
	client := NewMockRedisClient()
	ctx := context.Background()

	err := client.StoreChallenge(ctx, "test-token", "test-challenge", time.Minute)
	if err != nil {
		t.Errorf("StoreChallenge() error = %v", err)
	}

	err = client.DeleteChallenge(ctx, "test-token")
	if err != nil {
		t.Errorf("DeleteChallenge() error = %v", err)
	}

	_, err = client.GetChallenge(ctx, "test-token")
	if err == nil {
		t.Error("GetChallenge() error = nil, want error")
	}
}
