package middleware

import (
	"gateway/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ServiceAuthMiddleware(serverTokenService *service.ServerTokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("X-Service-Token")

		if authHeader == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "missing service header",
				},
			)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid service header",
				},
			)
			return
		}
		_, err := serverTokenService.ParseAccessServerToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid service header",
				},
			)
			return
		}

		c.Next()

	}
}
