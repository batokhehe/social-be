package categoryactivity

import (
	"net/http"
	"strconv"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"
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
// @Summary Create category activity
// @Description Create new category activity
// @Tags category-activity
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create category activity"
// @Success 200 {object} map[string]interface{}
// @Router /master/category-activities [post]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_501", "failed to create category activity"))
		return
	}

	response.Success(c, result)
}

// GetAll godoc
// @Summary Get all category activity
// @Description Get paginated list of category activity
// @Tags category-activity
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /master/category-activities [get]
func (h *Handler) GetAll(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	filters, appErr := query.ParseFilters(c.Request.URL.Query(), categoryActivityModel{})
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	items, meta, err := h.Service.GetPaginated(c.Request.Context(), page, filters)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_502", "failed to fetch category activity"))
		return
	}

	response.SuccessWithPagination(c, items, meta)
}

// GetByID godoc
// @Summary Get category activity by ID
// @Description Get detail category activity
// @Tags category-activity
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category Activity ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/category-activities/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_503", "category activity not found"))
		return
	}

	response.Success(c, item)
}

// Update godoc
// @Summary Update category activity
// @Description Update existing category activity
// @Tags category-activity
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category Activity ID"
// @Param request body UpdateRequest true "Update category activity"
// @Success 200 {object} map[string]interface{}
// @Router /master/category-activities/{id} [put]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_504", "failed to update category activity"))
		return
	}

	response.Success(c, item)
}

// Delete godoc
// @Summary Delete category activity
// @Description Soft delete category activity
// @Tags category-activity
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category Activity ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/category-activities/{id} [delete]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_505", "failed to delete category activity"))
		return
	}

	response.Success(c, gin.H{"message": "category activity deleted"})
}
