package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/coding-kelps/samarkand/internal/config"
)

type Server struct {
	cfg  *config.Config
	log  *slog.Logger
	grpc *grpc.Server
}

func New(cfg *config.Config, log *slog.Logger) *Server {
	s := &Server{cfg: cfg, log: log}
	s.grpc = s.newGRPCServer()
	return s
}

func (s *Server) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		addr := s.cfg.Server.Addr
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("grpc listen %s: %w", addr, err)
		}
		s.log.Info("gRPC server listening", "addr", addr)
		if err := s.grpc.Serve(lis); err != nil {
			return fmt.Errorf("grpc serve: %w", err)
		}
		return nil
	})

	return g.Wait()
}

func (s *Server) Stop() {
	s.log.Info("stopping gRPC server")
	s.grpc.GracefulStop()
}
