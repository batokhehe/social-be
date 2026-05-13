package validation

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"social-be/internal/pkg/apperror"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func BindJSON(c *gin.Context, req any) *apperror.AppError {
	if err := c.ShouldBindJSON(req); err != nil {
		return apperror.Wrap(err, http.StatusBadRequest, apperror.CodeInvalidBody, "invalid request body")
	}

	if err := validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return apperror.Validation(toFieldErrors(req, validationErrors))
		}

		return apperror.Wrap(err, http.StatusBadRequest, apperror.CodeValidationFailed, "validation failed")
	}

	return nil
}

func toFieldErrors(req any, validationErrors validator.ValidationErrors) []apperror.FieldError {
	out := make([]apperror.FieldError, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		out = append(out, apperror.FieldError{
			Field:   jsonFieldName(req, fieldErr.StructField()),
			Message: validationMessage(req, fieldErr),
		})
	}

	return out
}

func jsonFieldName(req any, structField string) string {
	t := reflect.TypeOf(req)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	field, ok := t.FieldByName(structField)
	if !ok {
		return strings.ToLower(structField)
	}

	jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
	if jsonName == "" || jsonName == "-" {
		return strings.ToLower(structField)
	}

	return jsonName
}

func validationMessage(req any, fieldErr validator.FieldError) string {
	field := jsonFieldName(req, fieldErr.StructField())

	switch fieldErr.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, fieldErr.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, fieldErr.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fieldErr.Param())
	default:
		return field + " is invalid"
	}
}
