package middleware

import (
	"github.com/gin-gonic/gin"
	"fmt"
)

func CORSMiddleware(allowedOrigin string) gin.HandlerFunc {
	if allowedOrigin == "" {
		allowedOrigin = "*" // fallback default
	}
	fmt.Println(allowedOrigin)
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
