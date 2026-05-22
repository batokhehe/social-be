package masterdonationcategory

import (
	"net/http"
	"strconv"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/validation"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func bindAndValidate(c *gin.Context, req interface{}) bool {
	if err := validation.BindJSON(c, req); err != nil {
		response.AbortError(c, err)
		return false
	}

	return true
}

// Create godoc
// @Summary Create master donation category
// @Description Create new master donation category
// @Tags master-donation-category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create master donation category"
// @Success 200 {object} map[string]interface{}
// @Router /master/donation-categories [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if ok := bindAndValidate(c, &req); !ok {
		return
	}

	actorID, ok := c.Get("user_id")
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "user not found"))
		return
	}

	ctx := c.Request.Context()

	result, err := h.Service.Create(ctx, req, actorID.(int))
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_321", "failed to create master donation category"))
		return
	}

	response.Success(c, result)
}

// GetAll godoc
// @Summary Get all master donation category
// @Description Get list of master donation category
// @Tags master-donation-category
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /master/donation-categories [get]
func (h *Handler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	items, meta, err := h.Service.GetPaginated(ctx, page)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_322", "failed to fetch master donation category"))
		return
	}

	response.SuccessWithPagination(c, items, meta)
}

// GetByID godoc
// @Summary Get master donation category by ID
// @Description Get detail master donation category
// @Tags master-donation-category
// @Produce json
// @Security BearerAuth
// @Param id path int true "Master Donation Category ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/donation-categories/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	ctx := c.Request.Context()

	item, err := h.Service.GetByID(ctx, id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_323", "master donation category not found"))
		return
	}

	response.Success(c, item)
}

// Update godoc
// @Summary Update master donation category
// @Description Update existing master donation category
// @Tags master-donation-category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Master Donation Category ID"
// @Param request body UpdateRequest true "Update master donation category"
// @Success 200 {object} map[string]interface{}
// @Router /master/donation-categories/{id} [put]
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

	actorID, ok := c.Get("user_id")
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "user not found"))
		return
	}

	ctx := c.Request.Context()

	item, err := h.Service.Update(ctx, id, req, actorID.(int))
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_324", "failed to update master donation category"))
		return
	}

	response.Success(c, item)
}

// Delete godoc
// @Summary Delete master donation category
// @Description Soft delete master donation category
// @Tags master-donation-category
// @Produce json
// @Security BearerAuth
// @Param id path int true "Master Donation Category ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/donation-categories/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	actorID, ok := c.Get("user_id")
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "user not found"))
		return
	}

	ctx := c.Request.Context()

	if err := h.Service.SoftDelete(ctx, id, actorID.(int)); err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_325", "failed to delete master donation category"))
		return
	}

	response.Success(c, gin.H{"message": "master donation category deleted"})
}
