package middleware

// import (
// 	"gateway/internal/service"
// 	"net/http"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// )

// func AuthMiddleware(tokenService *service.TokenService) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.Request.Header.Get("Authorization")

// 		if authHeader == "" {
// 			c.AbortWithStatusJSON(
// 				http.StatusUnauthorized,
// 				gin.H{
// 					"error": "missing authorization header",
// 				},
// 			)
// 			return
// 		}
// 		deviceHeader := c.Request.Header.Get("Device-ID")
// 		if deviceHeader == "" {
// 			c.AbortWithStatusJSON(
// 				http.StatusUnauthorized,
// 				gin.H{
// 					"error": "missing device-id header",
// 				},
// 			)
// 			return
// 		}
// 		parts := strings.Split(authHeader, " ")
// 		if len(parts) != 2 {
// 			c.AbortWithStatusJSON(
// 				http.StatusUnauthorized,
// 				gin.H{
// 					"error": "invalid authorization header",
// 				},
// 			)
// 			return
// 		}
// 		userId, err := tokenService.ParseAccessToken(parts[1], deviceHeader)
// 		if err != nil {
// 			c.AbortWithStatusJSON(
// 				http.StatusUnauthorized,
// 				gin.H{
// 					"error": "access denied",
// 				},
// 			)
// 			return
// 		}

// 		c.Set("user_id", userId)
// 		c.Set("device_id", deviceHeader)
// 		c.Next()

// 	}
// }
