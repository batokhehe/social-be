package middleware

import (
	"context"
	"net/http"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		if ctx.Err() == context.DeadlineExceeded && !c.Writer.Written() {
			response.WriteError(c, apperror.New(http.StatusGatewayTimeout, apperror.CodeRequestTimeout, "request timeout"))
		}
	}
}
