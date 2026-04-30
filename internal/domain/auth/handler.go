package auth

import (
	"social-be/internal/domain/user"
	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/security"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Handler struct {
	UserService *user.Service
}

var validate = validator.New()

// Register godoc
// @Summary Register user
// @Description Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body user.RegisterRequest true "Register request"
// @Success 200 {object} map[string]interface{}
// @Router /api/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req user.RegisterRequest

	// 🔹 Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Error("invalid register request", zap.Error(err))
		response.Error(c, "REQ_001", "invalid request body")
		return
	}

	// 🔹 Validate
	if err := validate.Struct(req); err != nil {
		logger.Log.Warn("validation failed", zap.Error(err))
		response.Error(c, "REQ_002", "validation failed")
		return
	}

	// 🔹 Service
	if err := h.UserService.Register(req.Name, req.Email, req.Password); err != nil {
		logger.Log.Error("register failed", zap.Error(err))
		response.Error(c, "SYS_001", "failed to register user")
		return
	}

	logger.Log.Info("user registered", zap.String("email", req.Email))

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

	// 🔹 Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Error("invalid login request", zap.Error(err))
		response.Error(c, "REQ_001", "invalid request body")
		return
	}

	// 🔹 Validate
	if err := validate.Struct(req); err != nil {
		logger.Log.Warn("login validation failed", zap.Error(err))
		response.Error(c, "REQ_002", "validation failed")
		return
	}

	// 🔹 Service
	accessToken, refreshToken, err := h.UserService.Login(req.Email, req.Password)
	if err != nil {
		response.Error(c, "AUTH_001", "invalid credentials")
		return
	}
	logger.Log.Info("login success", zap.String("email", req.Email))

	response.Success(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "REQ_001", "invalid request")
		return
	}

	claims, err := security.ParseToken(req.RefreshToken)
	if err != nil {
		response.Error(c, "AUTH_002", "invalid refresh token")
		return
	}

	if claims["type"] != "refresh" {
		response.Error(c, "AUTH_003", "invalid token type")
		return
	}

	userID := int(claims["user_id"].(float64))

	// ambil user dari DB
	user, _ := h.UserService.Repo.GetByID(userID)

	newAccess, _ := security.GenerateAccessToken(user.ID, user.Email, user.Role)

	response.Success(c, gin.H{
		"access_token": newAccess,
	})
}
