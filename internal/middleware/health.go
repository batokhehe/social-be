package middleware

import (
	"net/http"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/cache"
	"social-be/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func LivenessHandler(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok"})
}

func ReadinessHandler(c *gin.Context) {
	if _, err := cache.RDB.Ping(cache.Ctx).Result(); err != nil {
		response.WriteError(c, apperror.Wrap(err, http.StatusServiceUnavailable, "SYS_002", "service degraded"))
		return
	}

	response.Success(c, gin.H{"status": "ready"})
}
