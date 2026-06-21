package middleware

import (
	"bytes"
	"encoding/json"
	"gateway/internal/config"
	"gateway/internal/ratelimit"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ThrottleMiddleware(cfg *config.ThrottleConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	global := ratelimit.New(cfg.Limit, cfg.Window)

	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if !global.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}
		c.Next()
	}
}

func TelegramChatThrottleMiddleware(cfg *config.ThrottleConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	chatLimiter := ratelimit.New(cfg.ChatLimit, cfg.ChatWindow)

	return func(c *gin.Context) {
		raw, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(raw))

		key := "chat:ip:" + c.ClientIP()
		var payload struct {
			TelegramID string `json:"telegramID"`
		}
		if json.Unmarshal(raw, &payload) == nil && payload.TelegramID != "" {
			key = "chat:tg:" + payload.TelegramID
		}

		if !chatLimiter.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many chat requests",
			})
			return
		}
		c.Next()
	}
}

func AdminLoginThrottleMiddleware(cfg *config.ThrottleConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	loginLimiter := ratelimit.New(cfg.LoginLimit, cfg.LoginWindow)

	return func(c *gin.Context) {
		key := "login:ip:" + c.ClientIP()
		if !loginLimiter.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many login attempts",
			})
			return
		}
		c.Next()
	}
}
