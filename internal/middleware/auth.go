package middleware

import (
	"net/http"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/security"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeUnauthorized, "missing token"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeUnauthorized, "invalid token format"))
			return
		}

		claims, err := security.ParseToken(parts[1])
		if err != nil {
			response.AbortError(c, apperror.Wrap(err, http.StatusUnauthorized, apperror.CodeUnauthorized, "invalid token"))
			return
		}

		userID, err := security.GetIntClaim(claims, "user_id")
		if err != nil {
			response.AbortError(c, apperror.Wrap(err, http.StatusUnauthorized, apperror.CodeUnauthorized, "invalid token"))
			return
		}

		email, err := security.GetStringClaim(claims, "email")
		if err != nil {
			response.AbortError(c, apperror.Wrap(err, http.StatusUnauthorized, apperror.CodeUnauthorized, "invalid token"))
			return
		}

		role, err := security.GetIntClaim(claims, "role")
		if err != nil {
			response.AbortError(c, apperror.Wrap(err, http.StatusUnauthorized, apperror.CodeUnauthorized, "invalid token"))
			return
		}

		c.Set("user_id", userID)
		c.Set("email", email)
		c.Set("role", role)

		c.Next()
	}
}
