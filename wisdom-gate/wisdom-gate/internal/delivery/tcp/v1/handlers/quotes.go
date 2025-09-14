package handlers

import (
	"context"
	"fmt"
	"net"

	protocolUC "wisdom-gate/internal/application/protocol/usecase"
	quotesUC "wisdom-gate/internal/application/quotes/usecase"
	"wisdom-gate/internal/delivery/tcp/middleware"
)

type QuotesHandler struct {
	quotesStore *quotesUC.QuotesUseCase
}

func NewQuotesHandler(quotesStore *quotesUC.QuotesUseCase) *QuotesHandler {
	return &QuotesHandler{
		quotesStore: quotesStore,
	}
}

func (h *QuotesHandler) HandleQuoteRequest(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error {
	verified, ok := ctx.Value(middleware.VerifiedKey).(bool)
	if !ok || !verified {
		return fmt.Errorf("request not verified")
	}

	quote, err := h.quotesStore.GetRandomQuote(ctx)
	if err != nil {
		return err
	}

	quoteText := fmt.Sprintf("%s — %s", quote.Text, quote.Author)

	quoteMsg := &protocolUC.Message{
		Command: "QOT",
		Body:    quoteText,
	}

	if err := protocolUC.WriteMessage(conn, quoteMsg); err != nil {
		return fmt.Errorf("failed to send quote: %w", err)
	}

	return nil
}
