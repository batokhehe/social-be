package levelarea

import (
	"social-be/internal/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	Service *Service
}

var validate = validator.New()

func bindAndValidate(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, "REQ_001", "invalid request body")
		return false
	}

	if err := validate.Struct(req); err != nil {
		response.Error(c, "REQ_002", "validation failed")
		return false
	}

	return true
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if ok := bindAndValidate(c, &req); !ok {
		return
	}

	actorID, ok := c.Get("user_id")
	if !ok {
		response.Error(c, "AUTH_004", "user not found")
		return
	}

	result, err := h.Service.Create(req, actorID.(int))
	if err != nil {
		response.Error(c, "DB_101", "failed to create level area")
		return
	}

	response.Success(c, result)
}

func (h *Handler) GetAll(c *gin.Context) {
	items, err := h.Service.GetAll()
	if err != nil {
		response.Error(c, "DB_102", "failed to fetch level area")
		return
	}

	response.Success(c, items)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, "REQ_003", "invalid id")
		return
	}

	item, err := h.Service.GetByID(id)
	if err != nil {
		response.Error(c, "DB_103", "level area not found")
		return
	}

	response.Success(c, item)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, "REQ_003", "invalid id")
		return
	}

	var req UpdateRequest
	if ok := bindAndValidate(c, &req); !ok {
		return
	}

	actorID, ok := c.Get("user_id")
	if !ok {
		response.Error(c, "AUTH_004", "user not found")
		return
	}

	item, err := h.Service.Update(id, req, actorID.(int))
	if err != nil {
		response.Error(c, "DB_104", "failed to update level area")
		return
	}

	response.Success(c, item)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, "REQ_003", "invalid id")
		return
	}

	actorID, ok := c.Get("user_id")
	if !ok {
		response.Error(c, "AUTH_004", "user not found")
		return
	}

	if err := h.Service.SoftDelete(id, actorID.(int)); err != nil {
		response.Error(c, "DB_105", "failed to delete level area")
		return
	}

	response.Success(c, gin.H{"message": "level area deleted"})
}
