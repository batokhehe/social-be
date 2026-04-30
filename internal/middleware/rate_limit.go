package middleware

import (
	"net/http"
	"social-be/internal/pkg/cache"
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
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
