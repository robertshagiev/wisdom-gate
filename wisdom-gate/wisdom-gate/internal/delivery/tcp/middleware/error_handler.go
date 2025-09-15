package middleware

import (
	"context"
	"fmt"
	"net"

	"wisdom-gate/internal/application/protocol/consts"
	protocolUC "wisdom-gate/internal/application/protocol/usecase"
)

func ErrorHandlerMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error {
			err := next(ctx, conn, clientAddr, msg)
			if err != nil {
				errorMsg := &protocolUC.Message{
					Command: consts.CmdERR,
					Body:    fmt.Sprintf("ERROR: %s", err.Error()),
				}

				if writeErr := protocolUC.WriteMessage(conn, errorMsg); writeErr != nil {
					return fmt.Errorf("failed to send error response: %w", writeErr)
				}

				return err
			}
			return nil
		}
	}
}
