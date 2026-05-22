package masterdonaturgroup

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
// @Summary Create master donatur group
// @Description Create new master donatur group
// @Tags master-donatur-group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create master donatur group"
// @Success 200 {object} map[string]interface{}
// @Router /master/donatur-groups [post]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_311", "failed to create master donatur group"))
		return
	}

	response.Success(c, result)
}

// GetAll godoc
// @Summary Get all master donatur group
// @Description Get list of master donatur group
// @Tags master-donatur-group
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /master/donatur-groups [get]
func (h *Handler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	items, meta, err := h.Service.GetPaginated(ctx, page)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_312", "failed to fetch master donatur group"))
		return
	}

	response.SuccessWithPagination(c, items, meta)
}

// GetByID godoc
// @Summary Get master donatur group by ID
// @Description Get detail master donatur group
// @Tags master-donatur-group
// @Produce json
// @Security BearerAuth
// @Param id path int true "Master Donatur Group ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/donatur-groups/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	ctx := c.Request.Context()

	item, err := h.Service.GetByID(ctx, id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_313", "master donatur group not found"))
		return
	}

	response.Success(c, item)
}

// Update godoc
// @Summary Update master donatur group
// @Description Update existing master donatur group
// @Tags master-donatur-group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Master Donatur Group ID"
// @Param request body UpdateRequest true "Update master donatur group"
// @Success 200 {object} map[string]interface{}
// @Router /master/donatur-groups/{id} [put]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_314", "failed to update master donatur group"))
		return
	}

	response.Success(c, item)
}

// Delete godoc
// @Summary Delete master donatur group
// @Description Soft delete master donatur group
// @Tags master-donatur-group
// @Produce json
// @Security BearerAuth
// @Param id path int true "Master Donatur Group ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/donatur-groups/{id} [delete]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_315", "failed to delete master donatur group"))
		return
	}

	response.Success(c, gin.H{"message": "master donatur group deleted"})
}
