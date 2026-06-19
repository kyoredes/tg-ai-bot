package service

import (
	"context"
	"gateway/internal/client"
	"gateway/internal/dto"
	"gateway/internal/exceptions"
	"gateway/internal/logging"
	"time"

	authv1 "rageai/proto/gen/go/auth/v1"
	subscriptionv1 "rageai/proto/gen/go/subscription/v1"

	"go.uber.org/zap"
)

type TelegramService struct {
	clients *client.Clients
	timeout time.Duration
}

func NewTelegramService(clients *client.Clients, timeout time.Duration) *TelegramService {
	return &TelegramService{
		clients: clients,
		timeout: timeout,
	}
}

func (s *TelegramService) StartTelegram(telegramID string) (*dto.TelegramInfo, error) {
	logger := logging.Logger
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	authResp, err := s.clients.Auth.StartTelegram(ctx, &authv1.StartTelegramRequest{
		TelegramId: telegramID,
	})
	if err != nil {
		logger.Error("auth service grpc call failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	if authResp.GetUserId() == "" {
		logger.Error("auth service returned empty user_id")
		return nil, exceptions.ErrResponseExternalService
	}

	_, err = s.clients.Subscription.GetSubscriptionByUserId(ctx, &subscriptionv1.GetSubscriptionByUserIdRequest{
		UserId: authResp.GetUserId(),
	})
	if err != nil {
		logger.Error("subscription service grpc call failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	return &dto.TelegramInfo{
		TelegramID: telegramID,
		UserID:     authResp.GetUserId(),
		DeviceID:   telegramID,
	}, nil
}
