package dashboard

import (
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
