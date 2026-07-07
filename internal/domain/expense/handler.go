package expense

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/logger"
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

// expenseSortable whitelists ORDER BY columns (prevents SQL injection via sort).
var expenseSortable = map[string]string{
	"expense_date": "expense_date",
	"amount":       "amount",
	"created_at":   "created_at",
	"expense_no":   "expense_no",
	"status":       "status",
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

func parseExpenseSort(c *gin.Context) (ExpenseSort, *apperror.AppError) {
	sortParam := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort", "expense_date")))
	column, ok := expenseSortable[sortParam]
	if !ok {
		return ExpenseSort{}, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid sort column")
	}
	order := strings.ToLower(strings.TrimSpace(c.DefaultQuery("order", "desc")))
	if order != "asc" && order != "desc" {
		return ExpenseSort{}, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "order must be asc or desc")
	}
	return ExpenseSort{Column: column, Order: order}, nil
}

// mapExpenseError translates domain errors to HTTP responses.
func mapExpenseError(c *gin.Context, err error, code, message string) {
	switch {
	case errors.Is(err, ErrExpenseNotFound):
		response.AbortError(c, apperror.New(http.StatusNotFound, "DB_964", "expense not found"))
	case errors.Is(err, ErrCategoryNotFound):
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "category not found"))
	case errors.Is(err, ErrCategoryInactive):
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "category is inactive"))
	case errors.Is(err, ErrVolunteerNotFound):
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "volunteer not found"))
	case errors.Is(err, ErrInvalidExpenseDate):
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "expense_date is invalid"))
	default:
		// Log the real error server-side; return a generic message so raw DB
		// errors are never exposed to API consumers.
		logger.FromContext(c.Request.Context()).WithError(err).Error(message)
		response.AbortError(c, apperror.New(http.StatusInternalServerError, code, message))
	}
}

// GetAll godoc
// @Summary Get all expenses
// @Description Paginated list of expenses. Supports filter (status, category_id, volunteer_id, expense_date_from/to), search (expense_no, description) and sort (sort=expense_date|amount|created_at|expense_no|status, order=asc|desc).
// @Tags expense
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param sort query string false "Sort column"
// @Param order query string false "asc or desc"
// @Success 200 {object} map[string]interface{}
// @Router /expenses [get]
func (h *Handler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	filters, appErr := query.ParseFilters(c.Request.URL.Query(), expenseModel{})
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	sort, appErr := parseExpenseSort(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	items, meta, err := h.Service.GetPaginated(ctx, page, filters, sort)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Error("failed to fetch expenses")
		response.AbortError(c, apperror.New(http.StatusInternalServerError, "DB_962", "failed to fetch expenses"))
		return
	}
	response.SuccessWithPagination(c, items, meta)
}

// GetByID godoc
// @Summary Get expense by ID
// @Description Get expense detail with category, volunteer and audit actors
// @Tags expense
// @Produce json
// @Security BearerAuth
// @Param id path int true "Expense ID"
// @Success 200 {object} map[string]interface{}
// @Router /expenses/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		mapExpenseError(c, err, "DB_963", "failed to fetch expense")
		return
	}
	response.Success(c, item)
}

// Create godoc
// @Summary Create expense
// @Description Create a new expense (expense_no is auto-generated: EXP-YYYYMM-00001)
// @Tags expense
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create expense"
// @Success 200 {object} map[string]interface{}
// @Router /expenses [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if ok := bindAndValidate(c, &req); !ok {
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	item, err := h.Service.Create(c.Request.Context(), req, actor)
	if err != nil {
		mapExpenseError(c, err, "DB_961", "failed to create expense")
		return
	}
	response.Success(c, item)
}

// Update godoc
// @Summary Update expense
// @Description Update an existing expense (expense_no stays immutable)
// @Tags expense
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Expense ID"
// @Param request body UpdateRequest true "Update expense"
// @Success 200 {object} map[string]interface{}
// @Router /expenses/{id} [put]
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
		mapExpenseError(c, err, "DB_965", "failed to update expense")
		return
	}
	response.Success(c, item)
}

// Delete godoc
// @Summary Delete expense
// @Description Soft delete an expense (sets deleted_by)
// @Tags expense
// @Produce json
// @Security BearerAuth
// @Param id path int true "Expense ID"
// @Success 200 {object} map[string]interface{}
// @Router /expenses/{id} [delete]
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
	if err := h.Service.Delete(c.Request.Context(), id, actor); err != nil {
		mapExpenseError(c, err, "DB_966", "failed to delete expense")
		return
	}
	response.Success(c, gin.H{"message": "expense deleted"})
}
