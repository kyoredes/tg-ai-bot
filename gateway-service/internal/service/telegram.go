package service

import (
	"gateway/internal/config"
	"gateway/internal/dto"
	"gateway/internal/exceptions"
	"gateway/internal/logging"
	"gateway/internal/utils"
	"time"

	"go.uber.org/zap"
	"resty.dev/v3"
)

type TelegramService struct {
	authConfig  *config.AuthConfig
	subConfig   *config.SubConfig
	restyClient *resty.Client
}

func NewTelegramService(authConfig *config.AuthConfig, subConfig *config.SubConfig, restyClient *resty.Client) *TelegramService {
	restyClient.SetTimeout(time.Duration(authConfig.AuthTimeout) * time.Second)
	restyClient.SetRetryCount(authConfig.AuthRetryCount)
	return &TelegramService{
		authConfig:  authConfig,
		subConfig:   subConfig,
		restyClient: restyClient,
	}
}

func (s *TelegramService) StartTelegram(telegramID string) (*dto.TelegramInfo, error) {
	logger := logging.Logger
	restyClient := s.restyClient
	ch := make(chan dto.Response, 1)

	go utils.MakeRequest(
		restyClient,
		&dto.Request{
			Method: "POST",
			URL:    "http://" + s.authConfig.AuthHost + ":" + s.authConfig.AuthPort + "/telegram/start",
			Body:   []byte(`{"TelegramID": "` + telegramID + `"}`),
			Headers: map[string]string{
				"Authorization": "Bearer " + s.authConfig.AuthAPIKey,
				"Content-Type":  "application/json",
			},
			ExpectedStatusCode: 200,
		},
		ch,
	)
	responseRawAuth := <-ch
	raw, ok := responseRawAuth.Body["data"]
	if !ok || raw == nil {
		logger.Error("error response: empty data from auth service")
		return nil, exceptions.ErrResponseExternalService
	}

	data, ok := responseRawAuth.Body["data"].(map[string]any)
	if !ok {
		logger.Error("error respose: empty data from auth service")
		return nil, exceptions.ErrResponseExternalService
	}
	userId := data["user_id"].(string)
	go utils.MakeRequest(
		restyClient,
		&dto.Request{
			Method: "GET",
			URL:    "https://" + s.subConfig.SubHost + ":" + s.subConfig.SubPort + "/subscription/?user_id=" + userId,
			Headers: map[string]string{
				"Authorization": "Bearer " + s.subConfig.SubAPIKey,
				"Content-Type":  "application/json",
			},
			ExpectedStatusCode: 200,
		},
		ch,
	)
	responseRawSub := <-ch
	if !responseRawSub.Success {
		logger.Error("error response from sub service", zap.Int("status_code", responseRawSub.StatusCode))
		return nil, exceptions.ErrResponseExternalService
	}
	return &dto.TelegramInfo{
		TelegramID: telegramID,
	}, nil
}
