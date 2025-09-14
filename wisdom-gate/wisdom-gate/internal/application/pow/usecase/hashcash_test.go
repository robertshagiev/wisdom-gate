package usecase

import (
	"testing"
)

func TestVerifier_VerifySolution(t *testing.T) {
	verifier := NewVerifier()

	tests := []struct {
		name       string
		header     string
		difficulty int
		want       bool
		wantErr    bool
	}{
		{
			name:       "valid solution with difficulty 1",
			header:     "1:1:1234567890:127.0.0.1:8080:sha-256:test-nonce:1",
			difficulty: 1,
			want:       false,
			wantErr:    false,
		},
		{
			name:       "invalid solution",
			header:     "invalid-header",
			difficulty: 4,
			want:       false,
			wantErr:    false,
		},
		{
			name:       "empty header",
			header:     "",
			difficulty: 4,
			want:       false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := verifier.VerifySolution(tt.header, tt.difficulty)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verifier.VerifySolution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Verifier.VerifySolution() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNonceGenerator_GenerateNonce(t *testing.T) {
	generator := NewNonceGenerator()

	nonce, err := generator.GenerateNonce()
	if err != nil {
		t.Errorf("NonceGenerator.GenerateNonce() error = %v", err)
		return
	}

	if nonce == "" {
		t.Error("NonceGenerator.GenerateNonce() returned empty nonce")
	}

	nonce2, err := generator.GenerateNonce()
	if err != nil {
		t.Errorf("NonceGenerator.GenerateNonce() error = %v", err)
		return
	}

	if nonce == nonce2 {
		t.Error("NonceGenerator.GenerateNonce() returned same nonce twice")
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	if err != nil {
		t.Errorf("GenerateNonce() error = %v", err)
		return
	}

	if nonce == "" {
		t.Error("GenerateNonce() returned empty nonce")
	}
}
