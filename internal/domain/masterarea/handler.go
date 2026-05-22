package masterarea

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
// @Summary Create master area
// @Description Create new master area
// @Tags master-area
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create master area"
// @Success 200 {object} map[string]interface{}
// @Router /master/master-areas [post]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_201", "failed to create master area"))
		return
	}

	response.Success(c, result)
}

// GetAll godoc
// @Summary Get all master area
// @Description Get list of master area
// @Tags master-area
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /master/master-areas [get]
func (h *Handler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	items, meta, err := h.Service.GetPaginated(ctx, page)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_202", "failed to fetch master area"))
		return
	}

	response.SuccessWithPagination(c, items, meta)
}

// GetByID godoc
// @Summary Get master area by ID
// @Description Get detail master area
// @Tags master-area
// @Produce json
// @Security BearerAuth
// @Param id path int true "Master Area ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/master-areas/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	ctx := c.Request.Context()

	item, err := h.Service.GetByID(ctx, id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_203", "master area not found"))
		return
	}

	response.Success(c, item)
}

// Update godoc
// @Summary Update master area
// @Description Update existing master area
// @Tags master-area
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Master Area ID"
// @Param request body UpdateRequest true "Update master area"
// @Success 200 {object} map[string]interface{}
// @Router /master/master-areas/{id} [put]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_204", "failed to update master area"))
		return
	}

	response.Success(c, item)
}

// Delete godoc
// @Summary Delete master area
// @Description Soft delete master area
// @Tags master-area
// @Produce json
// @Security BearerAuth
// @Param id path int true "Master Area ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/master-areas/{id} [delete]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_205", "failed to delete master area"))
		return
	}

	response.Success(c, gin.H{"message": "master area deleted"})
}
