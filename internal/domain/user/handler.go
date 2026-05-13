package user

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
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
	ctx := c.Request.Context()

	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	users, meta, err := h.Service.GetPaginated(ctx, page)
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
// @Router /api/users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	ctx := c.Request.Context()

	user, err := h.Service.GetByID(ctx, id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_002", "user not found"))
		return
	}

	response.Success(c, user)
}

// UploadProfilePhoto godoc
// @Summary Upload profile photo
// @Description Upload profile photo to NAS and return URL
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param profile_photo formData file true "Profile photo"
// @Success 200 {object} map[string]interface{}
// @Router /api/users/upload-photo [post]
func (h *Handler) UploadProfilePhoto(c *gin.Context) {
	file, err := c.FormFile("profile_photo")
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidBody, "profile_photo is required"))
		return
	}

	// Validate file
	if err := validatePhotoFile(file); err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidBody, err.Error()))
		return
	}

	ctx := c.Request.Context()

	// Call NAS API to store image
	photoURL, err := h.Service.UploadPhotoToNAS(ctx, file)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to upload photo to NAS")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternal, "failed to upload photo"))
		return
	}

	logger.FromContext(c.Request.Context()).WithField("url", photoURL).Info("photo uploaded successfully")

	response.Success(c, gin.H{
		"profile_photo_url": photoURL,
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
