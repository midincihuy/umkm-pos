package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"umkm-pos/pkg/jwt"
	"umkm-pos/pkg/response"
)

// AuthMiddleware memvalidasi JWT yang diterbitkan Supabase Auth.
// Token dikirim frontend sebagai: Authorization: Bearer <supabase_access_token>
// User ID diambil dari claim "sub" (UUID standar Supabase).
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
			response.Error(c, http.StatusUnauthorized, "Token tidak valid atau sudah expired", nil)
			c.Abort()
			return
		}

		// Set userID ke context — diakses di handler via getUserID(c)
		c.Set("userID", userID)
		c.Next()
	}
}
