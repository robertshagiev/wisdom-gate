package usecase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

type Verifier struct{}

func NewVerifier() *Verifier {
	return &Verifier{}
}

func (v *Verifier) VerifySolution(header string, difficulty int) (bool, error) {
	hash := sha256.Sum256([]byte(header))
	hashStr := hex.EncodeToString(hash[:])

	requiredZeros := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hashStr, requiredZeros), nil
}

func GenerateNonce() (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	return base64.StdEncoding.EncodeToString(nonceBytes), nil
}
