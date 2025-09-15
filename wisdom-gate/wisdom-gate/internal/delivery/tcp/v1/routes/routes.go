package routes

import (
	"context"
	"fmt"
	"net"

	"wisdom-gate/internal/application/protocol/consts"
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
	case consts.CmdREQ:
		// REQ обрабатывается в middleware (PoWChallengeMiddleware)
		// Создаем пустой handler для middleware цепочки
		finalHandler = func(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error {
			return nil
		}
	case consts.CmdRES:
		finalHandler = handlers.QuotesHandler.HandleQuoteRequest
	case consts.CmdDISC:
		finalHandler = handlers.ConnectionHandler.HandleDisconnect
	default:
		return fmt.Errorf("unknown client command: %s", msg.Command)
	}

	handler := middlewareChain(finalHandler)

	return handler(ctx, conn, clientAddr, msg)
}
