package grpcserver

import (
	"auth/internal/config"
	"auth/internal/logging"
	"auth/internal/ratelimit"
	"fmt"
	"net"

	authv1 "agrobot/proto/gen/go/auth/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	cfg    *config.GRPCConfig
	server *grpc.Server
	lis    net.Listener
}

func NewServer(cfg *config.GRPCConfig, throttleCfg *config.ThrottleConfig, authServer *AuthServer) (*Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
	if err != nil {
		return nil, err
	}

	opts := []grpc.ServerOption{}
	if throttleCfg.Enabled {
		limiter := ratelimit.New(throttleCfg.Limit, throttleCfg.Window)
		opts = append(opts, grpc.UnaryInterceptor(ratelimit.UnaryServerInterceptor(limiter)))
	}

	s := grpc.NewServer(opts...)
	authv1.RegisterAuthServiceServer(s, authServer)

	return &Server{
		cfg:    cfg,
		server: s,
		lis:    lis,
	}, nil
}

func (s *Server) Start() error {
	logging.Logger.Info("gRPC server started",
		zap.String("host", s.cfg.Host),
		zap.String("port", s.cfg.Port),
	)
	return s.server.Serve(s.lis)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
