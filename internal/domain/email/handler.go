package email

import (
	"net/http"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/validation"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func bindAndValidate(c *gin.Context, req interface{}) bool {
	if err := validation.BindJSON(c, req); err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Warn("invalid request")
		response.AbortError(c, err)
		return false
	}

	return true
}

// @Summary Send async email check
// @Description Queues an email delivery request to SMTP and returns immediately. No real-time delivery confirmation is returned.
// @Tags email
// @Accept json
// @Produce json
// @Param request body SendRequest true "Email request"
// @Success 200 {object} map[string]interface{}
// @Router /email/test [post]
func (h *Handler) Send(c *gin.Context) {
	var req SendRequest
	if ok := bindAndValidate(c, &req); !ok {
		return
	}

	result, err := h.Service.Send(c.Request.Context(), req)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("dummy email send failed")
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternal, "failed to process dummy email request"))
		return
	}

	response.Success(c, result)
}
