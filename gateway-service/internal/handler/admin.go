package handler

import (
	"errors"
	"gateway/internal/dto"
	"gateway/internal/exceptions"
	"gateway/internal/logging"
	"gateway/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.adminService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "token": token})
}

func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.adminService.GetStats()
	if err != nil {
		logging.Logger.Error("GetStats failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "stats": stats})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	users, err := h.adminService.ListUsers(page, limit, search)
	if err != nil {
		logging.Logger.Error("ListUsers failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "users": users.Users, "total": users.Total})
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	user, err := h.adminService.GetUser(c.Param("id"))
	if err != nil {
		if errors.Is(err, exceptions.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "user": user})
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.adminService.UpdateUser(c.Param("id"), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "user": user})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	if err := h.adminService.DeleteUser(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *AdminHandler) ListSubscriptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	subs, err := h.adminService.ListSubscriptions(page, limit, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "subscriptions": subs.Subscriptions, "total": subs.Total})
}

func (h *AdminHandler) UpdateSubscription(c *gin.Context) {
	var req dto.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	sub, err := h.adminService.UpdateSubscription(c.Param("id"), req.StartsAt, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to update subscription"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "subscription": sub})
}

func (h *AdminHandler) DeleteSubscription(c *gin.Context) {
	if err := h.adminService.DeleteSubscription(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete subscription"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *AdminHandler) ListChatSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	sessions, err := h.adminService.ListChatSessions(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list chat sessions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "sessions": sessions.Sessions, "total": sessions.Total})
}

func (h *AdminHandler) GetChatHistory(c *gin.Context) {
	history, err := h.adminService.GetChatHistory(c.Param("telegramId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get chat history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "history": history})
}

func (h *AdminHandler) ClearChatHistory(c *gin.Context) {
	if err := h.adminService.ClearChatHistory(c.Param("telegramId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear chat history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *AdminHandler) ListProfileRoastSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	sessions, err := h.adminService.ListProfileRoastSessions(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list profile roast sessions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "sessions": sessions.Sessions, "total": sessions.Total})
}

func (h *AdminHandler) GetProfileRoastHistory(c *gin.Context) {
	history, err := h.adminService.GetProfileRoastHistory(c.Param("telegramId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get profile roast history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "history": history})
}

func (h *AdminHandler) ClearProfileRoastHistory(c *gin.Context) {
	if err := h.adminService.ClearProfileRoastHistory(c.Param("telegramId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear profile roast history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *AdminHandler) GetLLMConfig(c *gin.Context) {
	config, err := h.adminService.GetLLMConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get llm config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "config": config})
}

func (h *AdminHandler) GetSystemPrompt(c *gin.Context) {
	prompt, err := h.adminService.GetSystemPrompt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get system prompt"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "systemPrompt": prompt})
}

func (h *AdminHandler) UpdateSystemPrompt(c *gin.Context) {
	var req dto.UpdateSystemPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	prompt, err := h.adminService.UpdateSystemPrompt(req.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update system prompt"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "systemPrompt": prompt})
}

func (h *AdminHandler) GetServicesStatus(c *gin.Context) {
	servicesStatus, err := h.adminService.GetServicesStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get services status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "servicesStatus": servicesStatus})
}
