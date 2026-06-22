package service

import (
	"context"
	"errors"
	"gateway/internal/client"
	"gateway/internal/config"
	"gateway/internal/dto"
	"gateway/internal/exceptions"
	"gateway/internal/logging"
	"time"

	authv1 "agrobot/proto/gen/go/auth/v1"
	aiv1 "agrobot/proto/gen/go/ai/v1"
	subscriptionv1 "agrobot/proto/gen/go/subscription/v1"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminService struct {
	clients   *client.Clients
	adminCfg  *config.AdminConfig
	timeout   time.Duration
}

func NewAdminService(clients *client.Clients, adminCfg *config.AdminConfig, timeout time.Duration) *AdminService {
	return &AdminService{
		clients:  clients,
		adminCfg: adminCfg,
		timeout:  timeout,
	}
}

type adminClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (s *AdminService) Login(username, password string) (string, error) {
	if username != s.adminCfg.Username || password != s.adminCfg.Password {
		return "", errors.New("invalid credentials")
	}

	now := time.Now()
	claims := adminClaims{
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(8 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.adminCfg.JWTSecret))
}

func (s *AdminService) ValidateToken(tokenString string) error {
	token, err := jwt.ParseWithClaims(tokenString, &adminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.adminCfg.JWTSecret), nil
	})
	if err != nil {
		return err
	}
	claims, ok := token.Claims.(*adminClaims)
	if !ok || !token.Valid || claims.Role != "admin" {
		return errors.New("invalid token")
	}
	return nil
}

func (s *AdminService) GetStats() (*dto.AdminStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	var stats dto.AdminStats
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		resp, err := s.clients.Auth.GetStats(gctx, &authv1.GetAuthStatsRequest{})
		if err != nil {
			logging.Logger.Error("auth GetStats failed", zap.Error(err))
			return exceptions.ErrResponseExternalService
		}
		stats.Users.Total = resp.GetTotalUsers()
		stats.Users.New7d = resp.GetNewUsers_7D()
		return nil
	})

	g.Go(func() error {
		resp, err := s.clients.Subscription.GetStats(gctx, &subscriptionv1.GetSubscriptionStatsRequest{})
		if err != nil {
			logging.Logger.Error("subscription GetStats failed", zap.Error(err))
			return exceptions.ErrResponseExternalService
		}
		stats.Subscriptions.Total = resp.GetTotal()
		stats.Subscriptions.Active = resp.GetActive()
		stats.Subscriptions.Expired = resp.GetExpired()
		return nil
	})

	g.Go(func() error {
		resp, err := s.clients.AI.ListChatSessions(gctx, &aiv1.ListChatSessionsRequest{Page: 1, Limit: 1})
		if err != nil {
			logging.Logger.Error("ai ListChatSessions failed", zap.Error(err))
			return exceptions.ErrResponseExternalService
		}
		stats.Chat.Sessions = resp.GetTotal()
		return nil
	})

	g.Go(func() error {
		resp, err := s.clients.AI.ListProfileRoastSessions(gctx, &aiv1.ListProfileRoastSessionsRequest{Page: 1, Limit: 1})
		if err != nil {
			logging.Logger.Error("ai ListProfileRoastSessions failed", zap.Error(err))
			return exceptions.ErrResponseExternalService
		}
		stats.ProfileRoasts.Sessions = resp.GetTotal()
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return &stats, nil
}

func (s *AdminService) ListUsers(page, limit int, search string) (*dto.AdminUserList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.clients.Auth.ListUsers(ctx, &authv1.ListUsersRequest{
		Page:   int32(page),
		Limit:  int32(limit),
		Search: search,
	})
	if err != nil {
		logging.Logger.Error("auth ListUsers failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	users := make([]dto.AdminUser, len(resp.GetUsers()))
	for i, u := range resp.GetUsers() {
		users[i] = dto.AdminUser{
			UserID:     u.GetUserId(),
			Email:      u.GetEmail(),
			TelegramID: u.GetTelegramId(),
			CreatedAt:  u.GetCreatedAt(),
		}
	}

	return &dto.AdminUserList{Users: users, Total: resp.GetTotal()}, nil
}

func (s *AdminService) GetUser(userID string) (*dto.AdminUserDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.clients.Auth.GetUser(ctx, &authv1.GetUserRequest{UserId: userID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, exceptions.ErrUserNotFound
		}
		logging.Logger.Error("auth GetUser failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	return &dto.AdminUserDetail{
		UserID:     resp.GetUserId(),
		Email:      resp.GetEmail(),
		TelegramID: resp.GetTelegramId(),
		CreatedAt:  resp.GetCreatedAt(),
		UpdatedAt:  resp.GetUpdatedAt(),
	}, nil
}

func (s *AdminService) UpdateUser(userID, email string) (*dto.AdminUserDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	_, err := s.clients.Auth.UpdateUser(ctx, &authv1.UpdateUserRequest{
		UserId: userID,
		Email:  email,
	})
	if err != nil {
		logging.Logger.Error("auth UpdateUser failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	return s.GetUser(userID)
}

func (s *AdminService) DeleteUser(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	_, err := s.clients.Auth.DeleteUser(ctx, &authv1.DeleteUserRequest{UserId: userID})
	if err != nil {
		logging.Logger.Error("auth DeleteUser failed", zap.Error(err))
		return exceptions.ErrResponseExternalService
	}
	return nil
}

func (s *AdminService) ListSubscriptions(page, limit int, status string) (*dto.AdminSubscriptionList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.clients.Subscription.ListSubscriptions(ctx, &subscriptionv1.ListSubscriptionsRequest{
		Page:   int32(page),
		Limit:  int32(limit),
		Status: status,
	})
	if err != nil {
		logging.Logger.Error("subscription ListSubscriptions failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	subs := make([]dto.AdminSubscription, len(resp.GetSubscriptions()))
	for i, sub := range resp.GetSubscriptions() {
		subs[i] = dto.AdminSubscription{
			SubscriptionID: sub.GetSubscriptionId(),
			UserID:         sub.GetUserId(),
			StartsAt:       sub.GetStartsAt(),
			ExpiresAt:      sub.GetExpiresAt(),
		}
	}

	return &dto.AdminSubscriptionList{Subscriptions: subs, Total: resp.GetTotal()}, nil
}

func (s *AdminService) UpdateSubscription(subID string, startsAt, expiresAt int64) (*dto.AdminSubscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.clients.Subscription.UpdateSubscription(ctx, &subscriptionv1.UpdateSubscriptionRequest{
		SubscriptionId: subID,
		StartsAt:       startsAt,
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		logging.Logger.Error("subscription UpdateSubscription failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	return &dto.AdminSubscription{
		SubscriptionID: resp.GetSubscriptionId(),
		UserID:         resp.GetUserId(),
		StartsAt:       resp.GetStartsAt(),
		ExpiresAt:      resp.GetExpiresAt(),
	}, nil
}

func (s *AdminService) DeleteSubscription(subID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	_, err := s.clients.Subscription.DeleteSubscription(ctx, &subscriptionv1.DeleteSubscriptionRequest{
		SubscriptionId: subID,
	})
	if err != nil {
		logging.Logger.Error("subscription DeleteSubscription failed", zap.Error(err))
		return exceptions.ErrResponseExternalService
	}
	return nil
}

func (s *AdminService) ListChatSessions(page, limit int) (*dto.ChatSessionList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.clients.AI.ListChatSessions(ctx, &aiv1.ListChatSessionsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	})
	if err != nil {
		logging.Logger.Error("ai ListChatSessions failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	sessions := make([]dto.ChatSession, len(resp.GetSessions()))
	for i, s := range resp.GetSessions() {
		sessions[i] = dto.ChatSession{
			TelegramID:   s.GetTelegramId(),
			MessageCount: s.GetMessageCount(),
		}
	}

	return &dto.ChatSessionList{Sessions: sessions, Total: resp.GetTotal()}, nil
}

func (s *AdminService) GetChatHistory(telegramID string) (*dto.ChatHistory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.clients.AI.GetChatHistory(ctx, &aiv1.GetChatHistoryRequest{TelegramId: telegramID})
	if err != nil {
		logging.Logger.Error("ai GetChatHistory failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	messages := make([]dto.ChatMessage, len(resp.GetMessages()))
	for i, m := range resp.GetMessages() {
		messages[i] = dto.ChatMessage{Role: m.GetRole(), Content: m.GetContent()}
	}

	return &dto.ChatHistory{TelegramID: resp.GetTelegramId(), Messages: messages}, nil
}

func (s *AdminService) ClearChatHistory(telegramID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	_, err := s.clients.AI.ClearChatHistory(ctx, &aiv1.ClearChatHistoryRequest{TelegramId: telegramID})
	if err != nil {
		logging.Logger.Error("ai ClearChatHistory failed", zap.Error(err))
		return exceptions.ErrResponseExternalService
	}
	return nil
}

func (s *AdminService) ListProfileRoastSessions(page, limit int) (*dto.ProfileRoastSessionList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.clients.AI.ListProfileRoastSessions(ctx, &aiv1.ListProfileRoastSessionsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	})
	if err != nil {
		logging.Logger.Error("ai ListProfileRoastSessions failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	sessions := make([]dto.ProfileRoastSession, len(resp.GetSessions()))
	for i, session := range resp.GetSessions() {
		sessions[i] = dto.ProfileRoastSession{
			TelegramID: session.GetTelegramId(),
			RoastCount: session.GetRoastCount(),
		}
	}

	return &dto.ProfileRoastSessionList{Sessions: sessions, Total: resp.GetTotal()}, nil
}

func (s *AdminService) GetProfileRoastHistory(telegramID string) (*dto.ProfileRoastHistory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.clients.AI.GetProfileRoastHistory(ctx, &aiv1.GetProfileRoastHistoryRequest{TelegramId: telegramID})
	if err != nil {
		logging.Logger.Error("ai GetProfileRoastHistory failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	roasts := make([]dto.ProfileRoastItem, len(resp.GetRoasts()))
	for i, roast := range resp.GetRoasts() {
		roasts[i] = dto.ProfileRoastItem{
			CreatedAt:    roast.GetCreatedAt(),
			FirstName:    roast.GetFirstName(),
			LastName:     roast.GetLastName(),
			Username:     roast.GetUsername(),
			Bio:          roast.GetBio(),
			IsPremium:    roast.GetIsPremium(),
			LanguageCode: roast.GetLanguageCode(),
			HasPhoto:     roast.GetHasPhoto(),
			Response:     roast.GetResponse(),
		}
	}

	return &dto.ProfileRoastHistory{TelegramID: resp.GetTelegramId(), Roasts: roasts}, nil
}

func (s *AdminService) ClearProfileRoastHistory(telegramID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	_, err := s.clients.AI.ClearProfileRoastHistory(ctx, &aiv1.ClearProfileRoastHistoryRequest{TelegramId: telegramID})
	if err != nil {
		logging.Logger.Error("ai ClearProfileRoastHistory failed", zap.Error(err))
		return exceptions.ErrResponseExternalService
	}
	return nil
}

func (s *AdminService) GetLLMConfig() (*dto.LLMConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.clients.AI.GetLLMConfig(ctx, &aiv1.GetLLMConfigRequest{})
	if err != nil {
		logging.Logger.Error("ai GetLLMConfig failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	return &dto.LLMConfig{
		Model:       resp.GetModel(),
		Temperature: resp.GetTemperature(),
		MaxTokens:   resp.GetMaxTokens(),
		Debug:       resp.GetDebug(),
		Provider:    resp.GetProvider(),
		G4FModels:   resp.GetG4FModels(),
		UsesLiteLLM: resp.GetUsesLitellm(),
	}, nil
}

func (s *AdminService) GetSystemPrompt() (*dto.SystemPrompt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.clients.AI.GetSystemPrompt(ctx, &aiv1.GetSystemPromptRequest{})
	if err != nil {
		logging.Logger.Error("ai GetSystemPrompt failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	return &dto.SystemPrompt{
		Prompt:        resp.GetPrompt(),
		DefaultPrompt: resp.GetDefaultPrompt(),
		IsCustom:      resp.GetIsCustom(),
	}, nil
}

func (s *AdminService) UpdateSystemPrompt(prompt string) (*dto.SystemPrompt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.clients.AI.UpdateSystemPrompt(ctx, &aiv1.UpdateSystemPromptRequest{
		Prompt: prompt,
	})
	if err != nil {
		logging.Logger.Error("ai UpdateSystemPrompt failed", zap.Error(err))
		return nil, exceptions.ErrResponseExternalService
	}

	return &dto.SystemPrompt{
		Prompt:        resp.GetPrompt(),
		DefaultPrompt: resp.GetDefaultPrompt(),
		IsCustom:      resp.GetIsCustom(),
	}, nil
}
