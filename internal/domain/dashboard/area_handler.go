package dashboard

import (
	"errors"
	"net/http"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AreaHandler serves the Hu Ai and Xie Li dashboards. Each delegates to its own
// service; identity comes from the JWT (never a query param).
type AreaHandler struct {
	HuAi  *HuAiService
	XieLi *XieLiService
}

func NewAreaHandler(huAi *HuAiService, xieLi *XieLiService) *AreaHandler {
	return &AreaHandler{HuAi: huAi, XieLi: xieLi}
}

// GetHuAi godoc
// @Summary Hu Ai area dashboard
// @Description KPI summary + 6-month chart for the authenticated Hu Ai leader/deputy's area. Requires is_hu_ai_leader or is_hu_ai_deputy; otherwise 403.
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /dashboard/hu-ai [get]
func (h *AreaHandler) GetHuAi(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "user not found"))
		return
	}

	data, err := h.HuAi.GetDashboard(c.Request.Context(), userID)
	if err != nil {
		writeAreaError(c, err, "DB_954", "failed to fetch hu ai dashboard")
		return
	}
	response.Success(c, data)
}

// GetXieLi godoc
// @Summary Xie Li area dashboard
// @Description KPI summary + 6-month chart for the authenticated Xie Li leader/deputy's area. Requires is_xie_li_leader or is_xie_li_deputy; otherwise 403.
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /dashboard/xie-li [get]
func (h *AreaHandler) GetXieLi(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "user not found"))
		return
	}

	data, err := h.XieLi.GetDashboard(c.Request.Context(), userID)
	if err != nil {
		writeAreaError(c, err, "DB_955", "failed to fetch xie li dashboard")
		return
	}
	response.Success(c, data)
}

func writeAreaError(c *gin.Context, err error, code, message string) {
	switch {
	case errors.Is(err, ErrForbidden):
		response.AbortError(c, apperror.New(http.StatusForbidden, apperror.CodeForbidden, "you do not have access to this dashboard"))
	case errors.Is(err, ErrVolunteerNotFound):
		response.AbortError(c, apperror.New(http.StatusNotFound, apperror.CodeActorNotFound, "volunteer profile not found"))
	default:
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, code, message))
	}
}
