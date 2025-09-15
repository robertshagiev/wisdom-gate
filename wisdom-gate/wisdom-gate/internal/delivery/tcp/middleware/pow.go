package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"wisdom-gate/internal/adapters/redis"
	powUC "wisdom-gate/internal/application/pow/usecase"
	"wisdom-gate/internal/application/protocol/consts"
	protocolUC "wisdom-gate/internal/application/protocol/usecase"
	"wisdom-gate/internal/config"
)

func PoWChallengeMiddleware(redisClient redis.ClientInterface, cfg *config.Config) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error {
			if msg.Command != consts.CmdREQ {
				return next(ctx, conn, clientAddr, msg)
			}

			nonce, err := powUC.GenerateNonce()
			if err != nil {
				return fmt.Errorf("failed to generate nonce: %w", err)
			}

			expiresAt := time.Now().Add(cfg.Redis.ChallengeTTL).Unix()
			header := &protocolUC.HashcashHeader{
				Version:    1,
				Difficulty: cfg.POW.Difficulty,
				ExpiresAt:  expiresAt,
				Subject:    clientAddr,
				Algorithm:  consts.AlgorithmSha256,
				Nonce:      nonce,
			}

			token := nonce
			challengeStr := header.String()
			if err := redisClient.StoreChallenge(ctx, token, challengeStr, cfg.Redis.ChallengeTTL); err != nil {
				return fmt.Errorf("failed to store challenge: %w", err)
			}

			challengeMsg := &protocolUC.Message{
				Command: consts.CmdCHL,
				Body:    challengeStr,
			}

			if err := protocolUC.WriteMessage(conn, challengeMsg); err != nil {
				return fmt.Errorf("failed to send challenge: %w", err)
			}

			return nil
		}
	}
}

func PoWVerificationMiddleware(redisClient redis.ClientInterface, powVerifier powUC.VerifierInterface, cfg *config.Config, logger *slog.Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error {
			if msg.Command != consts.CmdREQ {
				return next(ctx, conn, clientAddr, msg)
			}

			header, err := protocolUC.ParseHashcashHeader(msg.Body)
			if err != nil {
				return fmt.Errorf("invalid header format: %w", err)
			}

			if header.IsExpired() {
				return fmt.Errorf("challenge expired")
			}

			if !header.ValidateSubject(clientAddr) {
				return fmt.Errorf("subject mismatch")
			}

			if header.Difficulty != cfg.POW.Difficulty {
				return fmt.Errorf("difficulty mismatch")
			}

			token := header.Nonce
			_, err = redisClient.GetChallenge(ctx, token)
			if err != nil {
				return fmt.Errorf("challenge not found")
			}

			spent, err := redisClient.MarkChallengeSpent(ctx, token, cfg.Redis.SpentTTL)
			if err != nil {
				return fmt.Errorf("failed to check replay: %w", err)
			}

			if !spent {
				return fmt.Errorf("challenge already used")
			}

			valid, err := powVerifier.VerifySolution(msg.Body, header.Difficulty)
			if err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}

			if !valid {
				return fmt.Errorf("insufficient proof of work")
			}

			if err := redisClient.DeleteChallenge(ctx, token); err != nil {
				logger.Warn("Failed to delete challenge from Redis", "token", token, "error", err)
			}

			ctx = context.WithValue(ctx, VerifiedKey, true)

			return next(ctx, conn, clientAddr, msg)
		}
	}
}
