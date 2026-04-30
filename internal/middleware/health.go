package middleware

import (
	"net/http"
	"social-be/internal/pkg/cache"

	"github.com/gin-gonic/gin"
)

func LivenessHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func ReadinessHandler(c *gin.Context) {
	if _, err := cache.RDB.Ping(cache.Ctx).Result(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
