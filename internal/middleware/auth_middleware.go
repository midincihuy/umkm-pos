package middleware

import (
	"net/http"
	"strings"
	"log"

	"github.com/gin-gonic/gin"
	"umkm-pos/pkg/jwt"
	"umkm-pos/pkg/response"
)

// AuthMiddleware memvalidasi JWT yang diterbitkan Google Auth.
// Token dikirim frontend sebagai: Authorization: Bearer <google_access_token>
// User ID diambil dari claim "sub" (UUID standar Google).
func AuthMiddleware(jwtHelper *jwt.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, "Token tidak ditemukan", nil)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := jwtHelper.GetUserID(tokenString)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Token tidak valid atau sudah expired coy", nil)
			log.Println(err.Error())
			c.Abort()
			return
		}

		// Set userID ke context — diakses di handler via getUserID(c)
		c.Set("userID", userID)
		c.Next()
	}
}
