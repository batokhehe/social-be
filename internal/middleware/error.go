package middleware

import (
	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			logger.Log.Error("request error", zap.Error(err))

			response.Error(c, "SYS_001", "internal server error")
		}
	}
}
