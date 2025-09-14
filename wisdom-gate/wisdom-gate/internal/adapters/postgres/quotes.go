package postgres

import (
	"context"
	"fmt"

	"wisdom-gate/internal/application/quotes/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

type QuotesRepository struct {
	db *pgxpool.Pool
}

func NewQuotesRepository(db *pgxpool.Pool) *QuotesRepository {
	return &QuotesRepository{db: db}
}

func (r *QuotesRepository) GetRandomQuote(ctx context.Context) (dto.Quote, error) {
	const op = "adapters.postgres.quotes.GetRandomQuote"

	query := `
		SELECT text, author 
		FROM quotes 
		ORDER BY RANDOM() 
		LIMIT 1
	`

	var quote dto.Quote
	err := r.db.QueryRow(ctx, query).Scan(&quote.Text, &quote.Author)
	if err != nil {
		return dto.Quote{}, fmt.Errorf("%s: failed to get random quote: %w", op, err)
	}

	return quote, nil
}
