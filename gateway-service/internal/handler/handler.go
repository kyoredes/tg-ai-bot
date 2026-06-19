package handler

import (
	"gateway/internal/service"
)

type Handler struct {
	Telegram *TelegramHandler
}

func NewHandler(
	telegramService *service.TelegramService,
) *Handler {
	return &Handler{
		Telegram: NewTelegramHandler(telegramService),
	}
}
