package middleware

import (
	"net/http"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Written() {
			return
		}

		if len(c.Errors) > 0 {
			appErr := apperror.From(c.Errors.Last().Err)

			logEntry := logger.FromContext(c.Request.Context()).
				WithField("code", appErr.Code).
				WithField("status", appErr.HTTPStatus)

			if appErr.HTTPStatus >= http.StatusInternalServerError {
				logEntry.WithError(appErr).Error("request error")
			} else {
				logEntry.WithError(appErr).Warn("request rejected")
			}

			response.WriteError(c, appErr)
		}
	}
}
