package middleware

import (
	"gateway/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DevAuthMiddleware(devCfg *config.DevConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "missing authorization header",
				},
			)
			return
		}
		if authHeader != devCfg.CommonPubKey {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid authorization header",
				},
			)
			return
		}
		c.Next()

	}
}
