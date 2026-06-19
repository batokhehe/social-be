package event

import (
	"net/http"
	"strconv"
	"time"

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
// @Summary Create event
// @Description Create new event by admin
// @Tags event
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create event"
// @Success 200 {object} map[string]interface{}
// @Router /master/events [post]
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
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to create event")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_900", "failed to create event"))
		return
	}
	response.Success(c, result)
}

// GetAll godoc
// @Summary Get all events
// @Description Get paginated list of events for admin
// @Tags event
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /master/events [get]
func (h *Handler) GetAll(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	items, meta, err := h.Service.GetAll(c.Request.Context(), page)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to fetch events")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_901", "failed to fetch events"))
		return
	}
	response.SuccessWithPagination(c, items, meta)
}

// GetByID godoc
// @Summary Get event by ID
// @Description Get event details for admin
// @Tags event
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/events/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_902", "event not found"))
		return
	}
	response.Success(c, item)
}

// Update godoc
// @Summary Update event
// @Description Update existing event by admin
// @Tags event
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Param request body UpdateRequest true "Update event"
// @Success 200 {object} map[string]interface{}
// @Router /master/events/{id} [put]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_903", "failed to update event"))
		return
	}
	response.Success(c, item)
}

// Delete godoc
// @Summary Delete event
// @Description Soft delete event by admin
// @Tags event
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/events/{id} [delete]
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
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_904", "failed to delete event"))
		return
	}
	response.Success(c, gin.H{"message": "event deleted"})
}

// CreateAttachment godoc
// @Summary Add attachment to event
// @Description Upload attachment file for a specific event
// @Tags event
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param attachment formData file true "Attachment file"
// @Param description formData string false "Attachment description"
// @Success 200 {object} map[string]interface{}
// @Router /master/events/{id}/attachments [post]
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
	desc := c.PostForm("description")
	actor, ok := actorID(c)
	if !ok {
		return
	}
	filePath, err := h.UploadService.UploadFile(c.Request.Context(), file, "event")
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to upload attachment")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternal, "failed to upload attachment"))
		return
	}
	attachment, err := h.Service.AddAttachment(c.Request.Context(), id, filePath, desc, actor)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_905", "failed to add event attachment"))
		return
	}
	response.Success(c, attachment)
}

// GetAttachments godoc
// @Summary Get event attachments
// @Description Get attachments for a specific event
// @Tags event
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} map[string]interface{}
// @Router /master/events/{id}/attachments [get]
func (h *Handler) GetAttachments(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	items, err := h.Service.GetAttachments(c.Request.Context(), id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_906", "failed to fetch event attachments"))
		return
	}
	response.Success(c, items)
}

// GetVolunteerEvents godoc
// @Summary Get active events for volunteer
// @Description List active events with valid end date
// @Tags event
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /events [get]
func (h *Handler) GetVolunteerEvents(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	items, meta, err := h.Service.GetActiveEvents(c.Request.Context(), actor, page)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to fetch volunteer events")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_907", "failed to fetch events"))
		return
	}
	response.SuccessWithPagination(c, items, meta)
}

// GetInvolvedEvents godoc
// @Summary Get events the current user is involved in
// @Description List events where the logged-in user is the PIC or a participant who has checked in
// @Tags event
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /events/involved [get]
func (h *Handler) GetInvolvedEvents(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	items, meta, err := h.Service.GetInvolvedEvents(c.Request.Context(), actor, page)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to fetch involved events")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_913", "failed to fetch involved events"))
		return
	}
	response.SuccessWithPagination(c, items, meta)
}

// GetAppliedVolunteerEvents godoc
// @Summary Get applied events for volunteer
// @Description List events that current volunteer has registered for
// @Tags event
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /events/applied [get]
func (h *Handler) GetAppliedVolunteerEvents(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	items, meta, err := h.Service.GetAppliedEvents(c.Request.Context(), actor, page)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to fetch applied volunteer events")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_912", "failed to fetch applied events"))
		return
	}
	response.SuccessWithPagination(c, items, meta)
}

// GetCompletedVolunteerEvents godoc
// @Summary Get completed events for volunteer
// @Description List events that current volunteer has completed
// @Tags event
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /events/completed [get]
func (h *Handler) GetCompletedVolunteerEvents(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	items, meta, err := h.Service.GetCompletedEvents(c.Request.Context(), actor, page)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to fetch completed volunteer events")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_912", "failed to fetch completed events"))
		return
	}
	response.SuccessWithPagination(c, items, meta)
}

// GetDetailEventsByVolunteer godoc
// @Summary Get detail events for volunteer
// @Description List events that current volunteer has registered for
// @Tags event
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /events/{id}/detail [get]
func (h *Handler) GetDetailEventsByVolunteer(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	items, meta, err := h.Service.GetDetailEventsByVolunteer(c.Request.Context(), actor, page)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("failed to fetch detail volunteer events")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_912", "failed to fetch detail events"))
		return
	}
	response.SuccessWithPagination(c, items, meta)
}

// GetVolunteerEventByID godoc
// @Summary Get active event detail for volunteer
// @Description Get event detail if active and still valid
// @Tags event
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} map[string]interface{}
// @Router /events/{id} [get]
func (h *Handler) GetVolunteerEventByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_908", "event not found"))
		return
	}
	if item.Status != "active" || item.EndAt.Before(time.Now()) {
		response.AbortError(c, apperror.New(http.StatusNotFound, apperror.CodeDatabase, "event not available"))
		return
	}
	item.Attendances = nil
	response.Success(c, item)
}

// ApplyToEvent godoc
// @Summary Apply to event
// @Description Register volunteer attendance for an event
// @Tags event
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} map[string]interface{}
// @Router /events/{id}/apply [post]
func (h *Handler) ApplyToEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	attendance, err := h.Service.ApplyToEvent(c.Request.Context(), id, actor, actor)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusBadRequest, "DB_909", "failed to apply event"))
		return
	}
	response.Success(c, attendance)
}

// CheckInEvent godoc
// @Summary Check in to event
// @Description Check in volunteer attendance with optional photo
// @Tags event
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param photo formData file false "Check-in photo"
// @Success 200 {object} map[string]interface{}
// @Router /events/{id}/checkin [post]
func (h *Handler) CheckInEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	var photoPath *string
	file, err := c.FormFile("photo")
	if err == nil {
		uploaded, uploadErr := h.UploadService.UploadFile(c.Request.Context(), file, "event")
		if uploadErr != nil {
			response.AbortError(c, apperror.Wrap(uploadErr, http.StatusInternalServerError, apperror.CodeInternal, "failed to upload photo"))
			return
		}
		photoPath = &uploaded
	}
	attendance, err := h.Service.CheckIn(c.Request.Context(), id, actor, photoPath, actor)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusBadRequest, "DB_910", "failed to check in"))
		return
	}
	response.Success(c, attendance)
}

// CheckOutEvent godoc
// @Summary Check out from event
// @Description Check out volunteer attendance with optional photo
// @Tags event
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param photo formData file false "Checkout photo"
// @Success 200 {object} map[string]interface{}
// @Router /events/{id}/checkout [post]
func (h *Handler) CheckOutEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}
	actor, ok := actorID(c)
	if !ok {
		return
	}
	var photoPath *string
	file, err := c.FormFile("photo")
	if err == nil {
		uploaded, uploadErr := h.UploadService.UploadFile(c.Request.Context(), file, "event")
		if uploadErr != nil {
			response.AbortError(c, apperror.Wrap(uploadErr, http.StatusInternalServerError, apperror.CodeInternal, "failed to upload photo"))
			return
		}
		photoPath = &uploaded
	}
	attendance, err := h.Service.CheckOut(c.Request.Context(), id, actor, photoPath, actor)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusBadRequest, "DB_911", "failed to check out"))
		return
	}
	response.Success(c, attendance)
}
