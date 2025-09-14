package usecase

import (
	"context"

	"wisdom-gate/internal/application/quotes/dto"
)

type repo interface {
	GetRandomQuote(ctx context.Context) (dto.Quote, error)
}
