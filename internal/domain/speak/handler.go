package speak

import (
	"net/http"
	"strconv"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/upload"
	"social-be/internal/pkg/validation"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service       *Service
	UploadService *upload.Service
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
// @Summary Create speak
// @Description Create a new speak
// @Tags speaks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create speak request"
// @Success 200 {object} map[string]interface{}
// @Router /speaks [post]
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
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to create speak")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_1100", "failed to create speak"))
		return
	}

	response.Success(c, result)
}

// GetAll godoc
// @Summary Get all speaks
// @Description Get paginated list of speaks
// @Tags speaks
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /speaks [get]
func (h *Handler) GetAll(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	items, meta, err := h.Service.GetAll(c.Request.Context(), page)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to fetch speaks")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_1101", "failed to fetch speaks"))
		return
	}

	response.SuccessWithPagination(c, items, meta)
}

// GetAllAsReporter godoc
// @Summary Get all as Reporter speaks
// @Description Get paginated list of speaks as Reporter
// @Tags speaks
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /speaks/reporter [get]
func (h *Handler) GetAllAsReporter(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	items, meta, err := h.Service.GetAllAsReporter(c.Request.Context(), page, actor)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to fetch speaks")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_1101", "failed to fetch speaks"))
		return
	}

	response.SuccessWithPagination(c, items, meta)
}

// GetAllAsRespondent godoc
// @Summary Get all as Respondent speaks
// @Description Get paginated list of speaks as Respondent
// @Tags speaks
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /speaks/respondent [get]
func (h *Handler) GetAllAsRespondent(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}

	items, meta, err := h.Service.GetAllAsRespondent(c.Request.Context(), page, actor)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to fetch speaks")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_1101", "failed to fetch speaks"))
		return
	}

	response.SuccessWithPagination(c, items, meta)
}

// GetByID godoc
// @Summary Get speak by ID
// @Description Get detail speak by ID
// @Tags speaks
// @Produce json
// @Security BearerAuth
// @Param id path int true "Speak ID"
// @Success 200 {object} map[string]interface{}
// @Router /speaks/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_1102", "speak not found"))
		return
	}

	response.Success(c, item)
}

// Update godoc
// @Summary Update speak
// @Description Update existing speak
// @Tags speaks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Speak ID"
// @Param request body UpdateRequest true "Update speak request"
// @Success 200 {object} map[string]interface{}
// @Router /speaks/{id} [put]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_1103", "failed to update speak"))
		return
	}

	response.Success(c, item)
}

// Delete godoc
// @Summary Delete speak
// @Description Soft delete speak
// @Tags speaks
// @Produce json
// @Security BearerAuth
// @Param id path int true "Speak ID"
// @Success 200 {object} map[string]interface{}
// @Router /speaks/{id} [delete]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_1104", "failed to delete speak"))
		return
	}

	response.Success(c, gin.H{"message": "speak deleted"})
}

// Action godoc
// @Summary Perform speak action
// @Description Execute action on speak
// @Tags speaks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Speak ID"
// @Param request body ActionRequest true "Action request"
// @Success 200 {object} map[string]interface{}
// @Router /speaks/{id}/action [post]
func (h *Handler) Action(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	var req ActionRequest
	if ok := bindAndValidate(c, &req); !ok {
		return
	}

	actor, ok := actorID(c)
	if !ok {
		return
	}

	item, err := h.Service.Action(c.Request.Context(), id, req, actor)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_1105", "failed to perform speak action"))
		return
	}

	response.Success(c, item)
}

// CreateAttachment godoc
// @Summary Upload speak attachment
// @Description Upload attachment file for a speak
// @Tags speaks
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Speak ID"
// @Param attachment formData file true "Attachment file"
// @Success 200 {object} map[string]interface{}
// @Router /speaks/{id}/attachments [post]
func (h *Handler) CreateAttachment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	file, err := c.FormFile("attachment")
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidBody, "attachment is required"))
		return
	}

	actor, ok := actorID(c)
	if !ok {
		return
	}

	filePath, err := h.UploadService.UploadFile(c.Request.Context(), file, "speak")
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to upload speak attachment")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternal, "failed to upload attachment"))
		return
	}

	item, err := h.Service.AddAttachment(c.Request.Context(), id, filePath, file.Filename, actor)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_1106", "failed to add speak attachment"))
		return
	}

	response.Success(c, item)
}

// GetAttachments godoc
// @Summary Get speak attachments
// @Description Get all attachments for a speak
// @Tags speaks
// @Produce json
// @Security BearerAuth
// @Param id path int true "Speak ID"
// @Success 200 {object} map[string]interface{}
// @Router /speaks/{id}/attachments [get]
func (h *Handler) GetAttachments(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	items, err := h.Service.GetAttachments(c.Request.Context(), id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_1107", "failed to fetch speak attachments"))
		return
	}

	response.Success(c, items)
}
