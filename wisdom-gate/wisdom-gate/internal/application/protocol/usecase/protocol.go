package usecase

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type HashcashHeader struct {
	Version    int
	Difficulty int
	ExpiresAt  int64
	Subject    string
	Algorithm  string
	Nonce      string
	Counter    int64
}

func (h *HashcashHeader) String() string {
	if h.Counter == 0 {
		return fmt.Sprintf("%d:%d:%d:%s:%s:%s",
			h.Version, h.Difficulty, h.ExpiresAt, h.Subject, h.Algorithm, h.Nonce)
	}
	return fmt.Sprintf("%d:%d:%d:%s:%s:%s:%s",
		h.Version, h.Difficulty, h.ExpiresAt, h.Subject, h.Algorithm, h.Nonce,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", h.Counter))))
}

func ParseHashcashHeader(header string) (*HashcashHeader, error) {
	parts := strings.Split(header, ":")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid header format")
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid version: %w", err)
	}

	difficulty, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid difficulty: %w", err)
	}

	expiresAt, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid expiresAt: %w", err)
	}

	// Subject может содержать IP:PORT, поэтому объединяем части
	var subject, algorithm, nonce string
	if len(parts) > 6 {
		// Если есть порт, объединяем IP:PORT
		subject = parts[3] + ":" + parts[4]
		algorithm = parts[5]
		nonce = parts[6]
	} else {
		subject = parts[3]
		algorithm = parts[4]
		nonce = parts[5]
	}

	h := &HashcashHeader{
		Version:    version,
		Difficulty: difficulty,
		ExpiresAt:  expiresAt,
		Subject:    subject,
		Algorithm:  algorithm,
		Nonce:      nonce,
	}

	// Counter есть только в solution (8+ частей)
	if len(parts) > 7 {
		counterBytes, err := base64.StdEncoding.DecodeString(parts[7])
		if err != nil {
			return nil, fmt.Errorf("invalid counter: %w", err)
		}
		counter, err := strconv.ParseInt(string(counterBytes), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid counter value: %w", err)
		}
		h.Counter = counter
	}

	return h, nil
}

func (h *HashcashHeader) IsExpired() bool {
	return time.Now().Unix() > h.ExpiresAt
}

func (h *HashcashHeader) ValidateSubject(expected string) bool {
	return h.Subject == expected
}
