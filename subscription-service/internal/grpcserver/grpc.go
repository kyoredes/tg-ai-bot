package grpcserver

import (
	"fmt"
	"net"
	"subscription/internal/config"
	"subscription/internal/logging"

	subscriptionv1 "rageai/proto/gen/go/subscription/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	cfg    *config.GRPCConfig
	server *grpc.Server
	lis    net.Listener
}

func NewServer(cfg *config.GRPCConfig, subscriptionServer *SubscriptionServer) (*Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer()
	subscriptionv1.RegisterSubscriptionServiceServer(s, subscriptionServer)

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
