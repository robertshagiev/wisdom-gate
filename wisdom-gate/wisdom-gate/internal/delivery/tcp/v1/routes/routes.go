package routes

import (
	"context"
	"fmt"
	"net"

	protocolUC "wisdom-gate/internal/application/protocol/usecase"
	"wisdom-gate/internal/delivery/tcp/middleware"
	"wisdom-gate/internal/delivery/tcp/v1/handlers"
)

func Route(
	ctx context.Context,
	conn net.Conn,
	clientAddr string,
	msg *protocolUC.Message,
	handlers handlers.Handlers,
	middlewareChain middleware.Middleware,
) error {
	var finalHandler middleware.Handler

	switch msg.Command {
	case "REQ":
		// REQ обрабатывается в middleware (PoWChallengeMiddleware)
		// Создаем пустой handler для middleware цепочки
		finalHandler = func(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error {
			return nil
		}
	case "RES":
		finalHandler = handlers.QuotesHandler.HandleQuoteRequest
	case "DISC":
		finalHandler = handlers.ConnectionHandler.HandleDisconnect
	default:
		return fmt.Errorf("unknown client command: %s", msg.Command)
	}

	handler := middlewareChain(finalHandler)

	return handler(ctx, conn, clientAddr, msg)
}
