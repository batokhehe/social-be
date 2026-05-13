package levelarea

import (
	"net/http"
	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/validation"
	"strconv"

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
// @Summary Create level area
// @Description Create new level area
// @Tags level-area
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create level area"
// @Success 200 {object} map[string]interface{}
// @Router /api/master/level-areas [post]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_101", "failed to create level area"))
		return
	}

	response.Success(c, result)
}

// GetAll godoc
// @Summary Get all level area
// @Description Get list of level area
// @Tags level-area
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/master/level-areas [get]
func (h *Handler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	items, meta, err := h.Service.GetPaginated(ctx, page)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_102", "failed to fetch level area"))
		return
	}

	response.SuccessWithPagination(c, items, meta)
}

// GetByID godoc
// @Summary Get level area by ID
// @Description Get detail level area
// @Tags level-area
// @Produce json
// @Security BearerAuth
// @Param id path int true "Level Area ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/master/level-areas/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	ctx := c.Request.Context()

	item, err := h.Service.GetByID(ctx, id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_103", "level area not found"))
		return
	}

	response.Success(c, item)
}

// Update godoc
// @Summary Update level area
// @Description Update existing level area
// @Tags level-area
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Level Area ID"
// @Param request body UpdateRequest true "Update level area"
// @Success 200 {object} map[string]interface{}
// @Router /api/master/level-areas/{id} [put]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_104", "failed to update level area"))
		return
	}

	response.Success(c, item)
}

// Delete godoc
// @Summary Delete level area
// @Description Soft delete level area
// @Tags level-area
// @Produce json
// @Security BearerAuth
// @Param id path int true "Level Area ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/master/level-areas/{id} [delete]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_105", "failed to delete level area"))
		return
	}

	response.Success(c, gin.H{"message": "level area deleted"})
}
