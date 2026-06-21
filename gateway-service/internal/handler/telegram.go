package handler

import (
	"errors"
	"gateway/internal/dto"
	"gateway/internal/exceptions"
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

func (h *TelegramHandler) GetProfile(c *gin.Context) {
	logger := logging.Logger

	var request dto.TelegramUserDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Debug("Wrong request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Wrong request",
		})
		return
	}

	profile, err := h.telegramService.GetProfile(request.TelegramID)
	if err != nil {
		if errors.Is(err, exceptions.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		logger.Error("Error getting telegram profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting telegram profile",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"profile": profile,
	})
}

func (h *TelegramHandler) GetSubscription(c *gin.Context) {
	logger := logging.Logger

	var request dto.TelegramUserDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Debug("Wrong request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Wrong request",
		})
		return
	}

	subscription, err := h.telegramService.GetSubscription(request.TelegramID)
	if err != nil {
		if errors.Is(err, exceptions.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		logger.Error("Error getting telegram subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error getting telegram subscription",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"subscription": subscription,
	})
}

func (h *TelegramHandler) Chat(c *gin.Context) {
	logger := logging.Logger

	var request dto.TelegramChatDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Debug("Wrong request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Wrong request",
		})
		return
	}

	chat, err := h.telegramService.Chat(request.TelegramID, request.Prompt)
	if err != nil {
		if errors.Is(err, exceptions.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		logger.Error("Error in telegram chat", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error processing chat request",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"chat":   chat,
	})
}

func (h *TelegramHandler) ClearChat(c *gin.Context) {
	logger := logging.Logger

	var request dto.TelegramUserDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Debug("Wrong request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Wrong request",
		})
		return
	}

	if err := h.telegramService.ClearChatHistory(request.TelegramID); err != nil {
		if errors.Is(err, exceptions.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		logger.Error("Error clearing telegram chat history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error clearing chat history",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
