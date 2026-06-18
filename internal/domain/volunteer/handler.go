package volunteer

import (
	"net/http"
	"strconv"
	"strings"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/validation"

	"github.com/gin-gonic/gin"
)

type Handler struct{ Service *Service }

func bindAndValidate(c *gin.Context, req interface{}) bool {
	if err := validation.BindJSON(c, req); err != nil {
		response.AbortError(c, err)
		return false
	}
	return true
}

func actorID(c *gin.Context) (int, bool) {
	value, ok := c.Get("user_id")
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "user not found"))
		return 0, false
	}
	id, ok := value.(int)
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "invalid user"))
		return 0, false
	}
	return id, true
}

// Create godoc
// @Summary Create volunteer
// @Description Create new volunteer
// @Tags volunteer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create volunteer"
// @Success 200 {object} map[string]interface{}
// @Router /master/volunteers [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if ok := bindAndValidate(c, &req); !ok {
		return
	}

	actor, ok := actorID(c)
	if !ok {
		return
	}

	result, err := h.Service.Create(c.Request.Context(), req, actor)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_801", "failed to create volunteer"))
		return
	}
	response.Success(c, result)
}

// GetAll godoc
// @Summary Get all volunteer
// @Description Get paginated list of volunteer
// @Tags volunteer
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param district query string false "Filter by district"
// @Param search query string false "Search by name or vis_id"
// @Success 200 {object} map[string]interface{}
// @Router /master/volunteers [get]
func (h *Handler) GetAll(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	district := strings.TrimSpace(c.Query("district"))
	search := strings.TrimSpace(c.Query("search"))
	items, meta, err := h.Service.GetPaginated(c.Request.Context(), page, district, search)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_802", "failed to fetch volunteer"))
		return
	}
	response.SuccessWithPagination(c, items, meta)
}

// GetDistricts godoc
// @Summary Get distinct volunteer districts
// @Description Get the list of distinct district values across volunteers
// @Tags volunteer
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /master/volunteers/districts [get]
func (h *Handler) GetDistricts(c *gin.Context) {
	items, err := h.Service.GetDistricts(c.Request.Context())
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_802", "failed to fetch volunteer districts"))
		return
	}
	response.Success(c, items)
}

// GetSelect godoc
// @Summary Get selected volunteer fields
// @Description Get paginated list of volunteer with selected fields
// @Tags volunteer
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /master/volunteers/select [get]
func (h *Handler) GetSelect(c *gin.Context) {
	items, meta, err := h.Service.GetSelect(c.Request.Context())
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_802", "failed to fetch volunteer"))
		return
	}
	response.SuccessWithPagination(c, items, meta)
}

func (h *Handler) VerifyResetPassword(c *gin.Context) {
	var req ResetPasswordVerifyRequest
	if ok := bindAndValidate(c, &req); !ok {
		return
	}

	result, err := h.Service.VerifyResetPassword(c.Request.Context(), req.Token)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).WithField("token", req.Token).Warn("invalid volunteer reset token")
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid or expired reset token"))
		return
	}

	response.Success(c, gin.H{
		"email":      result.Email,
		"valid":      result.Valid,
		"expires_at": result.ExpiresAt,
	})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if ok := bindAndValidate(c, &req); !ok {
		return
	}

	if err := h.Service.ResetPassword(c.Request.Context(), req.Token, req.Password); err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).WithField("token", req.Token).Warn("failed to reset volunteer password")
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid or expired reset token"))
		return
	}

	response.Success(c, gin.H{"message": "password updated"})
}

// GetByID godoc
// @Summary Get volunteer by ID
// @Description Get detail volunteer
// @Tags volunteer
// @Produce json
// @Security BearerAuth
// @Param id path int true "Volunteer ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/volunteers/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_803", "volunteer not found"))
		return
	}
	response.Success(c, item)
}

// Update godoc
// @Summary Update volunteer
// @Description Update existing volunteer
// @Tags volunteer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Volunteer ID"
// @Param request body UpdateRequest true "Update volunteer"
// @Success 200 {object} map[string]interface{}
// @Router /master/volunteers/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	var req UpdateRequest
	if ok := bindAndValidate(c, &req); !ok {
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	item, err := h.Service.Update(c.Request.Context(), id, req, actor)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_804", "failed to update volunteer"))
		return
	}
	response.Success(c, item)
}

// Delete godoc
// @Summary Delete volunteer
// @Description Soft delete volunteer
// @Tags volunteer
// @Produce json
// @Security BearerAuth
// @Param id path int true "Volunteer ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/volunteers/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	if err := h.Service.SoftDelete(c.Request.Context(), id, actor); err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_805", "failed to delete volunteer"))
		return
	}
	response.Success(c, gin.H{"message": "volunteer deleted"})
}
