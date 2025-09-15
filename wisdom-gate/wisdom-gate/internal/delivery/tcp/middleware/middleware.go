package middleware

import (
	"context"
	"net"
	"time"

	protocolUC "wisdom-gate/internal/application/protocol/usecase"
)

type Middleware func(next Handler) Handler

type Handler func(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error

type ContextKey string

const (
	ClientAddrKey ContextKey = "client_addr"
	VerifiedKey   ContextKey = "verified"
)

func Chain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}

		return next
	}
}

func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error {
			func() { _ = conn.SetReadDeadline(time.Now().Add(timeout)) }()
			func() { _ = conn.SetWriteDeadline(time.Now().Add(timeout)) }()

			return next(ctx, conn, clientAddr, msg)
		}
	}
}

func LoggingMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error {
			ctx = context.WithValue(ctx, ClientAddrKey, clientAddr)

			return next(ctx, conn, clientAddr, msg)
		}
	}
}
