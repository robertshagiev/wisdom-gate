package tcp

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"

	protocolUC "wisdom-gate/internal/application/protocol/usecase"
	v1 "wisdom-gate/internal/delivery/tcp/v1"
)

type Handler struct {
	api    *v1.API
	logger *slog.Logger
}

func NewHandler(api *v1.API, logger *slog.Logger) *Handler {
	return &Handler{
		api:    api,
		logger: logger,
	}
}

func (h *Handler) HandleConnection(ctx context.Context, conn net.Conn, clientAddr string) error {
	reader := bufio.NewReader(conn)

	for {
		msg, err := protocolUC.ReadMessage(reader)
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		h.logger.Debug("Received message", "command", msg.Command, "addr", clientAddr)

		if err := h.api.HandleMessage(ctx, conn, clientAddr, msg); err != nil {
			h.logger.Error("Error handling message", "addr", clientAddr, "error", err)
		}
	}
}
