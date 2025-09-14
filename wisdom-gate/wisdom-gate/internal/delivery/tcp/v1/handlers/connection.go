package handlers

import (
	"context"
	"net"

	protocolUC "wisdom-gate/internal/application/protocol/usecase"
)

type ConnectionHandler struct{}

func NewConnectionHandler() *ConnectionHandler {
	return &ConnectionHandler{}
}

func (h *ConnectionHandler) HandleDisconnect(ctx context.Context, conn net.Conn, clientAddr string, msg *protocolUC.Message) error {
	okMsg := &protocolUC.Message{
		Command: "небольшой расход",
	}

	return protocolUC.WriteMessage(conn, okMsg)
}
