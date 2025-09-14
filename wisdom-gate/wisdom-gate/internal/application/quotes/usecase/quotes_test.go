package usecase

import (
	"context"
	"testing"

	"wisdom-gate/internal/application/quotes/dto"
)

type MockQuotesRepository struct {
	quotes []dto.Quote
	err    error
}

func (m *MockQuotesRepository) GetRandomQuote(ctx context.Context) (dto.Quote, error) {
	if m.err != nil {
		return dto.Quote{}, m.err
	}
	if len(m.quotes) > 0 {
		return m.quotes[0], nil
	}
	return dto.Quote{}, nil
}

func TestQuotesUseCase_GetRandomQuote(t *testing.T) {
	tests := []struct {
		name     string
		mockRepo *MockQuotesRepository
		want     dto.Quote
		wantErr  bool
	}{
		{
			name: "successful quote retrieval",
			mockRepo: &MockQuotesRepository{
				quotes: []dto.Quote{
					{Text: "Test quote", Author: "Test Author"},
				},
				err: nil,
			},
			want:    dto.Quote{Text: "Test quote", Author: "Test Author"},
			wantErr: false,
		},
		{
			name: "repository error",
			mockRepo: &MockQuotesRepository{
				quotes: nil,
				err:    context.DeadlineExceeded,
			},
			want:    dto.Quote{},
			wantErr: true,
		},
		{
			name: "empty repository",
			mockRepo: &MockQuotesRepository{
				quotes: []dto.Quote{},
				err:    nil,
			},
			want:    dto.Quote{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewQuotesUseCase(tt.mockRepo)
			got, err := uc.GetRandomQuote(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("QuotesUseCase.GetRandomQuote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("QuotesUseCase.GetRandomQuote() = %v, want %v", got, tt.want)
			}
		})
	}
}
