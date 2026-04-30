package user

import (
	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	Service *Service
}

// GetUsers godoc
// @Summary Get all users
// @Description Get list of users
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	users, err := h.Service.GetUsers()
	if err != nil {
		logger.Log.Error("failed to get users", zap.Error(err))

		response.Error(c, "DB_001", "failed to fetch users")
		return
	}

	logger.Log.Info("get users success", zap.Int("count", len(users)))

	response.Success(c, users)
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get detail user by ID
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.Error(c, "REQ_003", "invalid id")
		return
	}

	user, err := h.Service.GetByID(id)
	if err != nil {
		response.Error(c, "DB_002", "user not found")
		return
	}

	response.Success(c, user)
}
