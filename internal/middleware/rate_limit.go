package middleware

import (
	"net/http"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/cache"
	"social-be/internal/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "rate:" + ip

		count, _ := cache.RDB.Incr(cache.Ctx, key).Result()

		if count == 1 {
			cache.RDB.Expire(cache.Ctx, key, time.Minute)
		}

		if count > 60 { // max 60 request / menit
			response.AbortError(c, apperror.New(http.StatusTooManyRequests, apperror.CodeRateLimited, "too many requests"))
			return
		}

		c.Next()
	}
}
