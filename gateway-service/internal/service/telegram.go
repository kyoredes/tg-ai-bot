package service

import (
	"context"
	"errors"
	"gateway/internal/client"
	"gateway/internal/dto"
	"gateway/internal/exceptions"
	"gateway/internal/logging"
	"time"

	authv1 "agrobot/proto/gen/go/auth/v1"
	aiv1 "agrobot/proto/gen/go/ai/v1"
	subscriptionv1 "agrobot/proto/gen/go/subscription/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *TelegramService) GetProfile(telegramID string) (*dto.TelegramProfile, error) {
	logger := logging.Logger
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	profile, err := s.fetchProfile(ctx, telegramID)
	if err == nil {
		return profile, nil
	}
	if !errors.Is(err, exceptions.ErrUserNotFound) {
		return nil, err
	}

	logger.Info("telegram profile not found, registering user", zap.String("telegram_id", telegramID))
	if _, err := s.StartTelegram(telegramID); err != nil {
		return nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.fetchProfile(ctx, telegramID)
}

func (s *TelegramService) fetchProfile(ctx context.Context, telegramID string) (*dto.TelegramProfile, error) {
	logger := logging.Logger

	authResp, err := s.clients.Auth.GetTelegramProfile(ctx, &authv1.GetTelegramProfileRequest{
		TelegramId: telegramID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, exceptions.ErrUserNotFound
		}
		logger.Error("auth service grpc GetTelegramProfile failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	return &dto.TelegramProfile{
		TelegramID: authResp.GetTelegramId(),
		UserID:     authResp.GetUserId(),
		Email:      authResp.GetEmail(),
	}, nil
}

func (s *TelegramService) GetSubscription(telegramID string) (*dto.TelegramSubscription, error) {
	logger := logging.Logger

	profile, err := s.GetProfile(telegramID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	subResp, err := s.clients.Subscription.GetSubscriptionByUserId(ctx, &subscriptionv1.GetSubscriptionByUserIdRequest{
		UserId: profile.UserID,
	})
	if err != nil {
		logger.Error("subscription service grpc call failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	return &dto.TelegramSubscription{
		SubscriptionID: subResp.GetSubscriptionId(),
		UserID:         subResp.GetUserId(),
		StartsAt:       subResp.GetStartsAt(),
		ExpiresAt:      subResp.GetExpiresAt(),
	}, nil
}

func (s *TelegramService) Chat(telegramID, prompt string) (*dto.TelegramChatResponse, error) {
	logger := logging.Logger

	if _, err := s.GetProfile(telegramID); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	aiResp, err := s.clients.AI.Chat(ctx, &aiv1.ChatRequest{
		TelegramId: telegramID,
		Prompt:     prompt,
	})
	if err != nil {
		logger.Error("ai service grpc Chat failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	return &dto.TelegramChatResponse{
		TelegramID: aiResp.GetTelegramId(),
		Response:   aiResp.GetResponse(),
	}, nil
}

func (s *TelegramService) ClearChatHistory(telegramID string) error {
	logger := logging.Logger

	if _, err := s.GetProfile(telegramID); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	_, err := s.clients.AI.ClearChatHistory(ctx, &aiv1.ClearChatHistoryRequest{
		TelegramId: telegramID,
	})
	if err != nil {
		logger.Error("ai service grpc ClearChatHistory failed", zap.Error(err))
		return exceptions.ErrResponseExternalService
	}
	return nil
}
