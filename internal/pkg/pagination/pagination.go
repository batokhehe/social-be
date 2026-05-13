package pagination

import (
	"math"
	"net/http"
	"strconv"

	"social-be/internal/pkg/apperror"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

type Query struct {
	Page   int
	Limit  int
	Offset int
}

type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func FromGin(c *gin.Context) (Query, *apperror.AppError) {
	page, err := parsePositiveInt(c.DefaultQuery("page", strconv.Itoa(defaultPage)), "page")
	if err != nil {
		return Query{}, err
	}

	limit, err := parsePositiveInt(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)), "limit")
	if err != nil {
		return Query{}, err
	}

	if limit > maxLimit {
		limit = maxLimit
	}

	return Query{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}, nil
}

func NewMeta(page, limit int, total int64) Meta {
	totalPages := 0
	if limit > 0 && total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    totalPages > 0 && page < totalPages,
		HasPrev:    page > 1 && totalPages > 0,
	}
}

func Slice[T any](items []T, q Query) ([]T, Meta) {
	total := len(items)
	if q.Offset >= total {
		return []T{}, NewMeta(q.Page, q.Limit, int64(total))
	}

	end := q.Offset + q.Limit
	if end > total {
		end = total
	}

	return items[q.Offset:end], NewMeta(q.Page, q.Limit, int64(total))
}

func parsePositiveInt(raw, field string) (int, *apperror.AppError) {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, &apperror.AppError{
			HTTPStatus: http.StatusBadRequest,
			Code:       apperror.CodeInvalidParam,
			Message:    "invalid query parameter",
			Detail: []apperror.FieldError{{
				Field:   field,
				Message: field + " must be a positive integer",
			}},
		}
	}

	return value, nil
}
