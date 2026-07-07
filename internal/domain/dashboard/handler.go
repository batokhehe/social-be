package dashboard

import (
	"errors"
	"net/http"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

// GetSummary godoc
// @Summary Dashboard summary KPIs
// @Description Returns dashboard KPI cards: total money donation, active donors and active volunteers (current vs previous month with percentage), plus the upcoming active event count.
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /dashboard/summary [get]
func (h *Handler) GetSummary(c *gin.Context) {
	summary, err := h.Service.GetSummary(c.Request.Context())
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_950", "failed to fetch dashboard summary"))
		return
	}

	response.Success(c, summary)
}

// GetHome godoc
// @Summary Dashboard homepage widgets
// @Description Returns homepage widgets: ongoing activities, latest donations, top volunteers (by contribution hours), and the impact summary (active volunteers, and completed activities = events whose end_at is in the past).
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /dashboard/home [get]
func (h *Handler) GetHome(c *gin.Context) {
	home, err := h.Service.GetHome(c.Request.Context())
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_951", "failed to fetch dashboard home"))
		return
	}

	response.Success(c, home)
}

// GetDonationsByCategory godoc
// @Summary Donations by category (pie chart)
// @Description Returns current-month money donations grouped by donation category, each with its summed amount, plus the grand total. Percentages are left to the frontend. Intended for a pie chart.
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /dashboard/donations-by-category [get]
func (h *Handler) GetDonationsByCategory(c *gin.Context) {
	data, err := h.Service.GetDonationsByCategory(c.Request.Context())
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_952", "failed to fetch donations by category"))
		return
	}

	response.Success(c, data)
}

// GetVolunteerDashboard godoc
// @Summary Personal volunteer dashboard
// @Description Statistics for the authenticated volunteer: current-month KPI cards (activity/philosophy/mission hours, my donors, my donations, collected donations) plus a 6-month chart. Volunteer identity is taken from the JWT, never from query params.
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /dashboard/volunteer [get]
func (h *Handler) GetVolunteerDashboard(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "user not found"))
		return
	}

	data, err := h.Service.GetVolunteerDashboard(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrVolunteerNotFound) {
			response.AbortError(c, apperror.New(http.StatusNotFound, apperror.CodeActorNotFound, "volunteer profile not found"))
			return
		}
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_953", "failed to fetch volunteer dashboard"))
		return
	}

	response.Success(c, data)
}

// currentUserID extracts the authenticated user id set by AuthMiddleware.
func currentUserID(c *gin.Context) (int, bool) {
	value, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	id, ok := value.(int)
	return id, ok
}
