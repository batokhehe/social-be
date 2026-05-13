package response

import (
	"net/http"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/pagination"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Success   bool   `json:"success"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Detail    any    `json:"detail,omitempty"`
	Meta      any    `json:"meta,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{
		Success:   true,
		Code:      apperror.CodeOK,
		Message:   "success",
		Data:      data,
		RequestID: RequestID(c),
	})
}

func SuccessWithPagination(c *gin.Context, data interface{}, meta pagination.Meta) {
	c.JSON(http.StatusOK, Envelope{
		Success:   true,
		Code:      apperror.CodeOK,
		Message:   "success",
		Data:      data,
		Meta:      meta,
		RequestID: RequestID(c),
	})
}

func Error(c *gin.Context, code, message string) {
	WriteError(c, apperror.New(http.StatusBadRequest, code, message))
}

func Abort(c *gin.Context, err error) {
	_ = c.Error(apperror.From(err))
	c.Abort()
}

func AbortError(c *gin.Context, err *apperror.AppError) {
	_ = c.Error(err)
	c.Abort()
}

func WriteError(c *gin.Context, err error) {
	appErr := apperror.From(err)
	status := appErr.HTTPStatus
	if status == 0 {
		status = http.StatusInternalServerError
	}

	c.JSON(status, Envelope{
		Success:   false,
		Code:      appErr.Code,
		Message:   appErr.Message,
		Detail:    appErr.Detail,
		RequestID: RequestID(c),
	})
}

func RequestID(c *gin.Context) string {
	if requestID, ok := c.Get("request_id"); ok {
		if value, ok := requestID.(string); ok {
			return value
		}
	}

	return c.GetHeader("X-Request-ID")
}
