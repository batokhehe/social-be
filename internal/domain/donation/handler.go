package donation

import (
	"net/http"
	"strconv"

	"social-be/internal/pkg/apperror"
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

func (h *Handler) GetAll(c *gin.Context) {
	items, err := h.Service.GetAll(c.Request.Context())
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_900", "failed to fetch donations"))
		return
	}
	response.Success(c, items)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := validation.BindJSON(c, &req); err != nil {
		response.AbortError(c, err)
		return
	}

	donation, err := h.Service.Create(c.Request.Context(), req)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_901", "failed to create donation"))
		return
	}

	response.Success(c, donation)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	donation, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_902", "donation not found"))
		return
	}

	response.Success(c, donation)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	var req UpdateRequest
	if err := validation.BindJSON(c, &req); err != nil {
		response.AbortError(c, err)
		return
	}

	donation, err := h.Service.Update(c.Request.Context(), id, req)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_903", "failed to update donation"))
		return
	}

	response.Success(c, donation)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	if err := h.Service.Delete(c.Request.Context(), id); err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_904", "failed to delete donation"))
		return
	}

	response.Success(c, gin.H{"message": "donation deleted"})
}
