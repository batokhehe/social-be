package user

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/upload"
	"social-be/internal/pkg/validation"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service       *Service
	UploadService *upload.Service
}

// GetUsers godoc
// @Summary Get all users
// @Description Get list of users
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	ctx := c.Request.Context()

	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	filters, appErr := query.ParseFilters(c.Request.URL.Query(), userModel{})
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	users, meta, err := h.Service.GetPaginated(ctx, page, filters)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to get users")

		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_001", "failed to fetch users"))
		return
	}

	logger.FromContext(c.Request.Context()).WithField("count", len(users)).Info("get users success")

	response.SuccessWithPagination(c, users, meta)
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get detail user by ID
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	roleVal, exists := c.Get("role")
	if !exists {
		response.AbortError(c, apperror.New(http.StatusForbidden, apperror.CodeForbidden, "role not found"))
		return
	}
	role, ok := roleVal.(int)
	if !ok {
		response.AbortError(c, apperror.New(http.StatusForbidden, apperror.CodeForbidden, "invalid role"))
		return
	}

	if role == 2 {
		userIDVal, ok := c.Get("user_id")
		if !ok {
			response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "user not found"))
			return
		}
		userID, ok := userIDVal.(int)
		if !ok || userID != id {
			response.AbortError(c, apperror.New(http.StatusForbidden, apperror.CodeForbidden, "forbidden"))
			return
		}
	}

	ctx := c.Request.Context()

	user, err := h.Service.GetByID(ctx, id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_002", "user not found"))
		return
	}

	response.Success(c, user)
}

// GetCurrentUser godoc
// @Summary Get current user profile
// @Description Get authenticated user profile
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /users/me [get]
func (h *Handler) GetCurrentUser(c *gin.Context) {
	userIDVal, ok := c.Get("user_id")
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "user not found"))
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "invalid user"))
		return
	}

	ctx := c.Request.Context()
	user, err := h.Service.GetByID(ctx, userID)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_002", "user not found"))
		return
	}

	response.Success(c, user)
}

// CreateUser godoc
// @Summary Create user
// @Description Create new user account by superadmin
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create user request"
// @Success 200 {object} map[string]interface{}
// @Router /admin/users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateRequest
	if err := validation.BindJSON(c, &req); err != nil {
		response.AbortError(c, err)
		return
	}

	ctx := c.Request.Context()
	if err := h.Service.Register(ctx, req); err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("create user failed")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternal, "failed to create user"))
		return
	}

	response.Success(c, gin.H{"message": "user created"})
}

// UploadProfilePhoto godoc
// @Summary Upload profile photo
// @Description Upload profile photo to NAS and update user profile photo
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param profile_photo formData file true "Profile photo"
// @Success 200 {object} map[string]interface{}
// @Router /users/upload-photo [post]
func (h *Handler) UploadProfilePhoto(c *gin.Context) {
	file, err := c.FormFile("profile_photo")
	if err != nil {
		response.AbortError(c, apperror.New(
			http.StatusBadRequest,
			apperror.CodeInvalidBody,
			"profile_photo is required",
		))
		return
	}

	if err := validatePhotoFile(file); err != nil {
		response.AbortError(c, apperror.New(
			http.StatusBadRequest,
			apperror.CodeInvalidBody,
			err.Error(),
		))
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		response.AbortError(c, apperror.New(
			http.StatusUnauthorized,
			apperror.CodeActorNotFound,
			"user not found",
		))
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		response.AbortError(c, apperror.New(
			http.StatusUnauthorized,
			apperror.CodeActorNotFound,
			"invalid user",
		))
		return
	}

	ctx := c.Request.Context()

	filePath, err := h.UploadService.UploadFile(ctx, file, "user")
	if err != nil {
		logger.FromContext(ctx).WithError(err).Error("failed to upload photo")

		response.AbortError(c, apperror.Wrap(
			err,
			http.StatusInternalServerError,
			apperror.CodeInternal,
			"failed to upload photo",
		))
		return
	}

	// update profile photo user
	if err := h.Service.UpdateProfilePhoto(ctx, userID, filePath); err != nil {
		logger.FromContext(ctx).WithError(err).Error("failed to update profile photo")

		response.AbortError(c, apperror.Wrap(
			err,
			http.StatusInternalServerError,
			"DB_003",
			"failed to update profile photo",
		))
		return
	}

	response.Success(c, gin.H{
		"profile_photo_path": filePath,
	})
}

func validatePhotoFile(file *multipart.FileHeader) error {
	const maxFileSize = 5 << 20 // 5MB
	if file.Size > maxFileSize {
		return fmt.Errorf("file size exceeds %d MB limit", maxFileSize/(1<<20))
	}

	// Check file extension
	allowedExt := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	ext := filepath.Ext(file.Filename)
	if !allowedExt[ext] {
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	return nil
}
