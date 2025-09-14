package tcp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"wisdom-gate/internal/adapters/postgres"
	"wisdom-gate/internal/adapters/redis"
	powUC "wisdom-gate/internal/application/pow/usecase"
	quotesUC "wisdom-gate/internal/application/quotes/usecase"
	"wisdom-gate/internal/config"
	"wisdom-gate/internal/delivery/tcp/middleware"
	v1 "wisdom-gate/internal/delivery/tcp/v1"
	"wisdom-gate/internal/delivery/tcp/v1/handlers"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	config   *config.Config
	logger   *slog.Logger
	handler  *Handler
	listener net.Listener
	wg       sync.WaitGroup
}

func NewServer(cfg *config.Config, logger *slog.Logger, db *pgxpool.Pool) (*Server, error) {
	redisClient, err := redis.NewClient(cfg.Redis.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	quoteRepo := postgres.NewQuotesRepository(db)
	powVerifier := powUC.NewVerifier()
	quotesUsecase := quotesUC.NewQuotesUseCase(quoteRepo)

	quotesHandler := handlers.NewQuotesHandler(quotesUsecase)
	connectionHandler := handlers.NewConnectionHandler()
	handlersCollection := handlers.NewHandlers(quotesHandler, connectionHandler)

	middlewareChain := middleware.Chain(
		middleware.TimeoutMiddleware(cfg.Server.ReadTimeout),
		middleware.LoggingMiddleware(),
		middleware.RateLimitMiddleware(middleware.NewRateLimiter(cfg.Server.RateLimit, cfg.Server.RateWindow)),
		middleware.PoWChallengeMiddleware(redisClient, cfg),
		middleware.PoWVerificationMiddleware(redisClient, powVerifier, cfg),
		middleware.ErrorHandlerMiddleware(),
	)

	api := v1.NewAPI(*handlersCollection, middlewareChain)
	handler := NewHandler(api, logger)

	return &Server{
		config:  cfg,
		logger:  logger,
		handler: handler,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	addr := ":" + s.config.Server.Port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	s.logger.Info("TCP wisdom-gate started", "addr", addr)

	go s.acceptConnections(ctx)

	<-ctx.Done()
	s.logger.Info("Shutting down TCP wisdom-gate...")

	if err := s.listener.Close(); err != nil {
		s.logger.Error("Failed to close listener", "error", err)
	}

	s.wg.Wait()
	s.logger.Info("TCP wisdom-gate stopped")
	return nil
}

func (s *Server) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					s.logger.Error("Failed to accept connection", "error", err)
					continue
				}
			}

			s.wg.Add(1)
			go s.handleConnection(ctx, conn)
		}
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	func() { _ = conn.SetReadDeadline(time.Now().Add(s.config.Server.ReadTimeout)) }()
	func() { _ = conn.SetWriteDeadline(time.Now().Add(s.config.Server.WriteTimeout)) }()

	clientAddr := conn.RemoteAddr().String()
	s.logger.Info("New connection", "addr", clientAddr)

	if err := s.handler.HandleConnection(ctx, conn, clientAddr); err != nil {
		s.logger.Error("Error handling client", "addr", clientAddr, "error", err)
	}

	s.logger.Info("Connection closed", "addr", clientAddr)
}
