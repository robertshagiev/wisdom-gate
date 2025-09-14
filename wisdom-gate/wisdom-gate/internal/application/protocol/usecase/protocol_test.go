package usecase

import (
	"testing"
	"time"
)

func TestHashcashHeader_String(t *testing.T) {
	tests := []struct {
		name   string
		header *HashcashHeader
		want   string
	}{
		{
			name: "challenge header without counter",
			header: &HashcashHeader{
				Version:    1,
				Difficulty: 4,
				ExpiresAt:  1234567890,
				Subject:    "127.0.0.1:8080",
				Algorithm:  "sha-256",
				Nonce:      "test-nonce",
				Counter:    0,
			},
			want: "1:4:1234567890:127.0.0.1:8080:sha-256:test-nonce",
		},
		{
			name: "solution header with counter",
			header: &HashcashHeader{
				Version:    1,
				Difficulty: 4,
				ExpiresAt:  1234567890,
				Subject:    "127.0.0.1:8080",
				Algorithm:  "sha-256",
				Nonce:      "test-nonce",
				Counter:    12345,
			},
			want: "1:4:1234567890:127.0.0.1:8080:sha-256:test-nonce:MTIzNDU=", // base64 encoded "12345"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.header.String(); got != tt.want {
				t.Errorf("HashcashHeader.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHashcashHeader(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    *HashcashHeader
		wantErr bool
	}{
		{
			name:   "valid challenge header",
			header: "1:4:1234567890:127.0.0.1:8080:sha-256:test-nonce",
			want: &HashcashHeader{
				Version:    1,
				Difficulty: 4,
				ExpiresAt:  1234567890,
				Subject:    "127.0.0.1:8080",
				Algorithm:  "sha-256",
				Nonce:      "test-nonce",
				Counter:    0,
			},
			wantErr: false,
		},
		{
			name:   "valid solution header with counter",
			header: "1:4:1234567890:127.0.0.1:8080:sha-256:test-nonce:MTIzNDU=",
			want: &HashcashHeader{
				Version:    1,
				Difficulty: 4,
				ExpiresAt:  1234567890,
				Subject:    "127.0.0.1:8080",
				Algorithm:  "sha-256",
				Nonce:      "test-nonce",
				Counter:    12345,
			},
			wantErr: false,
		},
		{
			name:    "invalid header format",
			header:  "invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty header",
			header:  "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid version",
			header:  "abc:4:1234567890:127.0.0.1:8080:sha-256:test-nonce",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid difficulty",
			header:  "1:abc:1234567890:127.0.0.1:8080:sha-256:test-nonce",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid expiresAt",
			header:  "1:4:abc:127.0.0.1:8080:sha-256:test-nonce",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHashcashHeader(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHashcashHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got == nil {
				t.Error("ParseHashcashHeader() returned nil")
				return
			}

			if *got != *tt.want {
				t.Errorf("ParseHashcashHeader() = %+v, want %+v", *got, *tt.want)
			}
		})
	}
}

func TestHashcashHeader_IsExpired(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name   string
		header *HashcashHeader
		want   bool
	}{
		{
			name: "expired header",
			header: &HashcashHeader{
				ExpiresAt: now - 1,
			},
			want: true,
		},
		{
			name: "valid header",
			header: &HashcashHeader{
				ExpiresAt: now + 60,
			},
			want: false,
		},
		{
			name: "header expiring now",
			header: &HashcashHeader{
				ExpiresAt: now - 1,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.header.IsExpired(); got != tt.want {
				t.Errorf("HashcashHeader.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashcashHeader_ValidateSubject(t *testing.T) {
	tests := []struct {
		name     string
		header   *HashcashHeader
		expected string
		want     bool
	}{
		{
			name: "valid subject",
			header: &HashcashHeader{
				Subject: "127.0.0.1:8080",
			},
			expected: "127.0.0.1:8080",
			want:     true,
		},
		{
			name: "invalid subject",
			header: &HashcashHeader{
				Subject: "127.0.0.1:8080",
			},
			expected: "192.168.1.1:8080",
			want:     false,
		},
		{
			name: "empty subject",
			header: &HashcashHeader{
				Subject: "",
			},
			expected: "",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.header.ValidateSubject(tt.expected); got != tt.want {
				t.Errorf("HashcashHeader.ValidateSubject() = %v, want %v", got, tt.want)
			}
		})
	}
}
