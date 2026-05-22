package middleware

import (
	"net/http"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func RoleMiddleware(allowedRoles ...int) gin.HandlerFunc {
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

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		response.AbortError(c, apperror.New(http.StatusForbidden, apperror.CodeForbidden, "forbidden"))
	}
}
