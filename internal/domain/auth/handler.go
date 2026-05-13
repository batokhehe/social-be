package auth

import (
	"net/http"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/validation"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func bindAndValidate(c *gin.Context, req interface{}) bool {
	if err := validation.BindJSON(c, req); err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Warn("invalid request")
		response.AbortError(c, err)
		return false
	}

	return true
}

// Register godoc
// @Summary Register user
// @Description Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 200 {object} map[string]interface{}
// @Router /api/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest

	if ok := bindAndValidate(c, &req); !ok {
		return
	}

	ctx := c.Request.Context()

	if err := h.Service.Register(ctx, req); err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("register failed")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternal, "failed to register user"))
		return
	}

	logger.FromContext(c.Request.Context()).WithField("email", req.Email).Info("user registered")

	response.Success(c, gin.H{
		"message": "user registered",
	})
}

// Login godoc
// @Summary Login user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} map[string]interface{}
// @Router /api/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if ok := bindAndValidate(c, &req); !ok {
		return
	}

	ctx := c.Request.Context()

	token, err := h.Service.Login(ctx, req)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusUnauthorized, apperror.CodeUnauthorized, "invalid credentials"))
		return
	}
	logger.FromContext(c.Request.Context()).WithField("email", req.Email).Info("login success")

	result := gin.H{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	}

	if token.User != nil {
		result["user"] = token.User
	}

	if token.Data != nil {
		result["data"] = token.Data
	}

	response.Success(c, result)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest

	if ok := bindAndValidate(c, &req); !ok {
		return
	}

	ctx := c.Request.Context()

	token, err := h.Service.Refresh(ctx, req)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("refresh failed")
		response.AbortError(c, apperror.Wrap(err, http.StatusUnauthorized, apperror.CodeRefreshFailed, "failed to refresh token"))
		return
	}

	response.Success(c, gin.H{
		"access_token": token.AccessToken,
	})
}
