package usecase

import (
	"context"

	"wisdom-gate/internal/application/quotes/dto"
)

type QuotesUseCase struct {
	repo repoInterface
}

func NewQuotesUseCase(repo repoInterface) *QuotesUseCase {

	return &QuotesUseCase{repo: repo}
}

func (s *QuotesUseCase) GetRandomQuote(ctx context.Context) (dto.Quote, error) {
	return s.repo.GetRandomQuote(ctx)
}
