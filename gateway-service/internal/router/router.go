package router

import (
	"gateway/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	h *handler.Handler,
	authMiddleware gin.HandlerFunc,
	adminAuthMiddleware gin.HandlerFunc,
	corsMiddleware gin.HandlerFunc,
) *gin.Engine {
	router := gin.Default()
	router.Use(corsMiddleware)

	tg := router.Group("/telegram")
	tg.Use(authMiddleware)
	tg.POST("/start", h.Telegram.StartTelegram)
	tg.POST("/profile", h.Telegram.GetProfile)
	tg.POST("/subscription", h.Telegram.GetSubscription)
	tg.POST("/chat", h.Telegram.Chat)
	tg.POST("/chat/clear", h.Telegram.ClearChat)

	admin := router.Group("/admin")
	admin.POST("/login", h.Admin.Login)

	protected := admin.Group("")
	protected.Use(adminAuthMiddleware)
	protected.GET("/stats", h.Admin.GetStats)
	protected.GET("/services", h.Admin.GetServicesStatus)
	protected.GET("/users", h.Admin.ListUsers)
	protected.GET("/users/:id", h.Admin.GetUser)
	protected.PATCH("/users/:id", h.Admin.UpdateUser)
	protected.DELETE("/users/:id", h.Admin.DeleteUser)
	protected.GET("/subscriptions", h.Admin.ListSubscriptions)
	protected.PATCH("/subscriptions/:id", h.Admin.UpdateSubscription)
	protected.DELETE("/subscriptions/:id", h.Admin.DeleteSubscription)

	chat := protected.Group("/chat")
	chat.GET("/sessions", h.Admin.ListChatSessions)
	chat.GET("/history/:telegramId", h.Admin.GetChatHistory)
	chat.DELETE("/history/:telegramId", h.Admin.ClearChatHistory)

	llm := protected.Group("/llm")
	llm.GET("", h.Admin.GetLLMConfig)
	llm.GET("/prompt", h.Admin.GetSystemPrompt)
	llm.PATCH("/prompt", h.Admin.UpdateSystemPrompt)

	return router
}
