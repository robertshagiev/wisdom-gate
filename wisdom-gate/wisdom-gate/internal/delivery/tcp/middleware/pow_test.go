package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"wisdom-gate/internal/adapters/redis"
	powUC "wisdom-gate/internal/application/pow/usecase"
	"wisdom-gate/internal/application/protocol/usecase"
	"wisdom-gate/internal/config"
)

func TestPoWChallengeMiddleware(t *testing.T) {
	mockRedis := redis.NewMockRedisClient()

	cfg := &config.Config{
		Redis: config.RedisConfig{
			ChallengeTTL: 20 * time.Second,
		},
		POW: config.POWConfig{
			Difficulty: 4,
		},
	}

	middleware := PoWChallengeMiddleware(mockRedis, cfg)

	nextHandler := func(ctx context.Context, conn net.Conn, clientAddr string, msg *usecase.Message) error {
		return nil
	}

	handler := middleware(nextHandler)

	conn := &mockConn{}

	reqMsg := &usecase.Message{
		Command: "REQ",
		Body:    "",
	}

	err := handler(context.Background(), conn, "127.0.0.1:8080", reqMsg)
	if err != nil {
		t.Errorf("PoWChallengeMiddleware() error = %v", err)
	}

	if len(conn.writtenData) == 0 {
		t.Error("No data was written to connection")
	}

	// Проверяем что challenge сохранился в Redis
	// Поскольку мы не знаем nonce, проверим что в Redis есть хотя бы один challenge
}

func TestPoWVerificationMiddleware(t *testing.T) {
	mockRedis := redis.NewMockRedisClient()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	verifier := powUC.NewVerifier()

	cfg := &config.Config{
		Redis: config.RedisConfig{
			SpentTTL: 2 * time.Minute,
		},
		POW: config.POWConfig{
			Difficulty: 1, // Низкая сложность для тестов
		},
	}

	middleware := PoWVerificationMiddleware(mockRedis, verifier, cfg, logger)

	nextHandler := func(ctx context.Context, conn net.Conn, clientAddr string, msg *usecase.Message) error {
		return nil
	}

	handler := middleware(nextHandler)

	conn := &mockConn{}

	challenge := &usecase.HashcashHeader{
		Version:    1,
		Difficulty: 1,
		ExpiresAt:  time.Now().Add(time.Minute).Unix(),
		Subject:    "127.0.0.1:8080",
		Algorithm:  "sha-256",
		Nonce:      "test-nonce",
		Counter:    0,
	}

	challengeStr := challenge.String()
	err := mockRedis.StoreChallenge(context.Background(), challenge.Nonce, challengeStr, time.Minute)
	if err != nil {
		t.Errorf("Failed to store challenge: %v", err)
	}

	// Создаем валидное решение (с низкой сложностью)
	// Нужно найти counter который даст hash с нужным количеством нулей
	var validCounter int64
	for counter := int64(0); counter < 1000; counter++ {
		solution := &usecase.HashcashHeader{
			Version:    1,
			Difficulty: 1,
			ExpiresAt:  time.Now().Add(time.Minute).Unix(),
			Subject:    "127.0.0.1:8080",
			Algorithm:  "sha-256",
			Nonce:      "test-nonce",
			Counter:    counter,
		}

		// Проверяем валидность решения
		hash := sha256.Sum256([]byte(solution.String()))
		hashStr := hex.EncodeToString(hash[:])
		if strings.HasPrefix(hashStr, "0") {
			validCounter = counter
			break
		}
	}

	solution := &usecase.HashcashHeader{
		Version:    1,
		Difficulty: 1,
		ExpiresAt:  time.Now().Add(time.Minute).Unix(),
		Subject:    "127.0.0.1:8080",
		Algorithm:  "sha-256",
		Nonce:      "test-nonce",
		Counter:    validCounter,
	}

	solutionStr := solution.String()

	// Тестируем RES команду
	resMsg := &usecase.Message{
		Command: "RES",
		Body:    solutionStr,
	}

	err = handler(context.Background(), conn, "127.0.0.1:8080", resMsg)
	if err != nil {
		t.Errorf("PoWVerificationMiddleware() error = %v", err)
	}
}

// mockConn - мок для net.Conn
type mockConn struct {
	writtenData [][]byte
}

func (m *mockConn) Read(b []byte) (n int, err error) { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error) {
	m.writtenData = append(m.writtenData, b)
	return len(b), nil
}
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }
