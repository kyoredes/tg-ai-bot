package grpcserver

import (
	"context"
	"subscription/internal/logging"
	"subscription/internal/service"

	subscriptionv1 "rageai/proto/gen/go/subscription/v1"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SubscriptionServer struct {
	subscriptionv1.UnimplementedSubscriptionServiceServer
	service *service.SubscriptionService
}

func NewSubscriptionServer(service *service.SubscriptionService) *SubscriptionServer {
	return &SubscriptionServer{service: service}
}

func (s *SubscriptionServer) GetSubscriptionByUserId(ctx context.Context, req *subscriptionv1.GetSubscriptionByUserIdRequest) (*subscriptionv1.GetSubscriptionByUserIdResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	sub, err := s.service.EnsureSubByUserId(userID)
	if err != nil {
		logging.Logger.Error("grpc GetSubscriptionByUserId failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get subscription")
	}

	return &subscriptionv1.GetSubscriptionByUserIdResponse{
		SubscriptionId: sub.Uuid.String(),
		UserId:         sub.UserID.String(),
		StartsAt:       sub.StartsAt.Unix(),
		ExpiresAt:      sub.ExpiresAt.Unix(),
	}, nil
}
