package client

import (
	"fmt"
	"gateway/internal/config"

	authv1 "rageai/proto/gen/go/auth/v1"
	subscriptionv1 "rageai/proto/gen/go/subscription/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Auth         authv1.AuthServiceClient
	Subscription subscriptionv1.SubscriptionServiceClient
	authConn     *grpc.ClientConn
	subConn      *grpc.ClientConn
}

func NewClients(authCfg *config.AuthConfig, subCfg *config.SubConfig) (*Clients, error) {
	authConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%s", authCfg.AuthHost, authCfg.AuthGRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("auth grpc dial: %w", err)
	}

	subConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%s", subCfg.SubHost, subCfg.SubGRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		authConn.Close()
		return nil, fmt.Errorf("subscription grpc dial: %w", err)
	}

	return &Clients{
		Auth:         authv1.NewAuthServiceClient(authConn),
		Subscription: subscriptionv1.NewSubscriptionServiceClient(subConn),
		authConn:     authConn,
		subConn:      subConn,
	}, nil
}

func (c *Clients) Close() error {
	if err := c.authConn.Close(); err != nil {
		return err
	}
	return c.subConn.Close()
}
