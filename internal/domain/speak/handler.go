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
