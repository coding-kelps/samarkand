package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Server struct {
	addr   string
	logger *slog.Logger
	grpc   *grpc.Server
}

type ServerConfig struct {
	Addr   string
	Logger *slog.Logger
}

func NewServer(cfg *ServerConfig) *Server {
	return &Server{
		addr:   cfg.Addr,
		grpc:   newGRPCServer(),
		logger: cfg.Logger,
	}
}

func (s *Server) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		addr := s.addr

		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("grpc listen %s: %w", addr, err)
		}

		s.logger.Info("gRPC server listening", "addr", addr)
		if err := s.grpc.Serve(lis); err != nil {
			return fmt.Errorf("grpc serve: %w", err)
		}

		return nil
	})

	return g.Wait()
}

func (s *Server) Stop() {
	s.logger.Info("stopping gRPC server")
	s.grpc.GracefulStop()
}
