package masterexpensecategory

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

// sortable whitelists ORDER BY columns (prevents ORDER BY injection).
var sortable = map[string]string{
	"id":         "id",
	"code":       "code",
	"name":       "name",
	"created_at": "created_at",
}

func bindAndValidate(c *gin.Context, req interface{}) bool {
	if err := validation.BindJSON(c, req); err != nil {
		response.AbortError(c, err)
		return false
	}
	return true
}

func parseSort(c *gin.Context) (Sort, *apperror.AppError) {
	sortParam := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort", "created_at")))
	column, ok := sortable[sortParam]
	if !ok {
		return Sort{}, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid sort column")
	}
	order := strings.ToLower(strings.TrimSpace(c.DefaultQuery("order", "desc")))
	if order != "asc" && order != "desc" {
		return Sort{}, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "order must be asc or desc")
	}
	return Sort{Column: column, Order: order}, nil
}

func mapError(c *gin.Context, err error, code, message string) {
	switch {
	case errors.Is(err, ErrCategoryNotFound):
		response.AbortError(c, apperror.New(http.StatusNotFound, "MEC_404", "expense category not found"))
	case errors.Is(err, ErrCodeExists):
		response.AbortError(c, apperror.New(http.StatusConflict, "MEC_409", "code already exists"))
	case errors.Is(err, ErrNameExists):
		response.AbortError(c, apperror.New(http.StatusConflict, "MEC_409", "name already exists"))
	default:
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, code, message))
	}
}

// GetAll godoc
// @Summary Get all expense categories
// @Description Paginated list. Search: code, name. Filter: active. Sort: sort=id|code|name|created_at, order=asc|desc (default created_at desc).
// @Tags master-expense-category
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param sort query string false "Sort column"
// @Param order query string false "asc or desc"
// @Param active query bool false "Filter by active"
// @Success 200 {object} map[string]interface{}
// @Router /master-expense-categories [get]
func (h *Handler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	filters, appErr := query.ParseFilters(c.Request.URL.Query(), masterExpenseCategoryModel{})
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	sort, appErr := parseSort(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	items, meta, err := h.Service.GetPaginated(ctx, page, filters, sort)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "MEC_002", "failed to fetch expense categories"))
		return
	}
	response.SuccessWithPagination(c, items, meta)
}

// GetSelect godoc
// @Summary Get active expense categories for dropdown
// @Description Returns only active, non-deleted categories (id, code, name)
// @Tags master-expense-category
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /master-expense-categories/select [get]
func (h *Handler) GetSelect(c *gin.Context) {
	items, err := h.Service.GetSelect(c.Request.Context())
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "MEC_003", "failed to fetch expense categories"))
		return
	}
	response.Success(c, items)
}

// GetByID godoc
// @Summary Get expense category by ID
// @Description Get expense category detail
// @Tags master-expense-category
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]interface{}
// @Router /master-expense-categories/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		mapError(c, err, "MEC_004", "failed to fetch expense category")
		return
	}
	response.Success(c, item)
}

// Create godoc
// @Summary Create expense category
// @Description Create a new expense category (code and name must be unique)
// @Tags master-expense-category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create expense category"
// @Success 200 {object} map[string]interface{}
// @Router /master-expense-categories [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if ok := bindAndValidate(c, &req); !ok {
		return
	}
	item, err := h.Service.Create(c.Request.Context(), req)
	if err != nil {
		mapError(c, err, "MEC_001", "failed to create expense category")
		return
	}
	response.Success(c, item)
}

// Update godoc
// @Summary Update expense category
// @Description Update an expense category (code and name must be unique)
// @Tags master-expense-category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param request body UpdateRequest true "Update expense category"
// @Success 200 {object} map[string]interface{}
// @Router /master-expense-categories/{id} [put]
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
	item, err := h.Service.Update(c.Request.Context(), id, req)
	if err != nil {
		mapError(c, err, "MEC_005", "failed to update expense category")
		return
	}
	response.Success(c, item)
}

// Delete godoc
// @Summary Delete expense category
// @Description Soft delete an expense category
// @Tags master-expense-category
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]interface{}
// @Router /master-expense-categories/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	if err := h.Service.Delete(c.Request.Context(), id); err != nil {
		mapError(c, err, "MEC_006", "failed to delete expense category")
		return
	}
	response.Success(c, gin.H{"message": "expense category deleted"})
}
