package handlers

type Handlers struct {
	QuotesHandler     *QuotesHandler
	ConnectionHandler *ConnectionHandler
}

func NewHandlers(quotesHandler *QuotesHandler, connectionHandler *ConnectionHandler) *Handlers {
	return &Handlers{
		QuotesHandler:     quotesHandler,
		ConnectionHandler: connectionHandler,
	}
}
