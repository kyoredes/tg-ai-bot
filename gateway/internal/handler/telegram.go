package handler

import (
	"gateway/internal/dto"
	"gateway/internal/logging"
	"gateway/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TelegramHandler struct {
	telegramService *service.TelegramService
}

func NewTelegramHandler(telegramService *service.TelegramService) *TelegramHandler {
	return &TelegramHandler{
		telegramService: telegramService,
	}
}

func (h *TelegramHandler) StartTelegram(c *gin.Context) {
	logger := logging.Logger

	var request dto.TelegramUserDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Debug("Wrong request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Wrong request",
		})
		return
	}

	info, err := h.telegramService.StartTelegram(request.TelegramID)
	if err != nil {
		logger.Error("Error starting telegram", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error starting telegram",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"info":   info,
	})
}
