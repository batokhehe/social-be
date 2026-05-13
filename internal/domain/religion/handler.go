package religion

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
// @Summary Create religion
// @Description Create new religion
// @Tags religion
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create religion"
// @Success 200 {object} map[string]interface{}
// @Router /api/master/religions [post]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_601", "failed to create religion"))
		return
	}
	response.Success(c, result)
}

// GetAll godoc
// @Summary Get all religion
// @Description Get paginated list of religion
// @Tags religion
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /api/master/religions [get]
func (h *Handler) GetAll(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	items, meta, err := h.Service.GetPaginated(c.Request.Context(), page)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_602", "failed to fetch religion"))
		return
	}
	response.SuccessWithPagination(c, items, meta)
}

// GetByID godoc
// @Summary Get religion by ID
// @Description Get detail religion
// @Tags religion
// @Produce json
// @Security BearerAuth
// @Param id path int true "Religion ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/master/religions/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_603", "religion not found"))
		return
	}
	response.Success(c, item)
}

// Update godoc
// @Summary Update religion
// @Description Update existing religion
// @Tags religion
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Religion ID"
// @Param request body UpdateRequest true "Update religion"
// @Success 200 {object} map[string]interface{}
// @Router /api/master/religions/{id} [put]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_604", "failed to update religion"))
		return
	}
	response.Success(c, item)
}

// Delete godoc
// @Summary Delete religion
// @Description Soft delete religion
// @Tags religion
// @Produce json
// @Security BearerAuth
// @Param id path int true "Religion ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/master/religions/{id} [delete]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_605", "failed to delete religion"))
		return
	}
	response.Success(c, gin.H{"message": "religion deleted"})
}
