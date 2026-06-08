package donation

import (
	"net/http"
	"strconv"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/validation"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

// GetAll godoc
// @Summary Get all donations
// @Description Get list of donations
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /donations [get]
func (h *Handler) GetAll(c *gin.Context) {
	items, err := h.Service.GetAll(c.Request.Context())
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_900", "failed to fetch donations"))
		return
	}
	response.Success(c, items)
}

// Create godoc
// @Summary Create donation
// @Description Create new donation
// @Tags donation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create donation"
// @Success 200 {object} map[string]interface{}
// @Router /donations [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := validation.BindJSON(c, &req); err != nil {
		response.AbortError(c, err)
		return
	}

	donation, err := h.Service.Create(c.Request.Context(), req)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_901", "failed to create donation"))
		return
	}

	response.Success(c, donation)
}

// GetByID godoc
// @Summary Get donation by ID
// @Description Get detail donation
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Param id path int true "Donation ID"
// @Success 200 {object} map[string]interface{}
// @Router /donations/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	donation, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_902", "donation not found"))
		return
	}

	response.Success(c, donation)
}

// Update godoc
// @Summary Update donation
// @Description Update existing donation
// @Tags donation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Donation ID"
// @Param request body UpdateRequest true "Update donation"
// @Success 200 {object} map[string]interface{}
// @Router /donations/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	var req UpdateRequest
	if err := validation.BindJSON(c, &req); err != nil {
		response.AbortError(c, err)
		return
	}

	donation, err := h.Service.Update(c.Request.Context(), id, req)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_903", "failed to update donation"))
		return
	}

	response.Success(c, donation)
}

// Delete godoc
// @Summary Delete donation
// @Description Delete donation
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Param id path int true "Donation ID"
// @Success 200 {object} map[string]interface{}
// @Router /donations/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	if err := h.Service.Delete(c.Request.Context(), id); err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_904", "failed to delete donation"))
		return
	}

	response.Success(c, gin.H{"message": "donation deleted"})
}
