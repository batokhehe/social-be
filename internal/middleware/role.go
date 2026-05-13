package middleware

import (
	"net/http"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func RoleMiddleware(requiredRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			response.AbortError(c, apperror.New(http.StatusForbidden, apperror.CodeForbidden, "role not found"))
			return
		}

		role, ok := roleVal.(int)
		if !ok {
			response.AbortError(c, apperror.New(http.StatusForbidden, apperror.CodeForbidden, "invalid role"))
			return
		}

		if role != requiredRole {
			response.AbortError(c, apperror.New(http.StatusForbidden, apperror.CodeForbidden, "forbidden"))
			return
		}

		c.Next()
	}
}
