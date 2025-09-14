package v1

import (
	"context"
	"net"

	protocolUC "wisdom-gate/internal/application/protocol/usecase"
	"wisdom-gate/internal/delivery/tcp/middleware"
	"wisdom-gate/internal/delivery/tcp/v1/handlers"
	clientRoutesV1 "wisdom-gate/internal/delivery/tcp/v1/routes"
)

type API struct {
	handlers        handlers.Handlers
	middlewareChain middleware.Middleware
}

func NewAPI(handlers handlers.Handlers, middlewareChain middleware.Middleware) *API {
	return &API{
		handlers:        handlers,
		middlewareChain: middlewareChain,
	}
}

func (api *API) HandleMessage(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error {
	return clientRoutesV1.Route(ctx, conn, clientAddr, msg, api.handlers, api.middlewareChain)
}
