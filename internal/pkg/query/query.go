package query

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"social-be/internal/pkg/apperror"

	"gorm.io/gorm"
)

type DateRange struct {
	From *time.Time
	To   *time.Time
}

type Filters struct {
	Equals map[string]any
	Like   map[string]string
	Range  map[string]DateRange
}

var reservedQueryParams = map[string]struct{}{
	"page":  {},
	"limit": {},
}

func ParseFilters(values url.Values, example interface{}) (Filters, *apperror.AppError) {
	filters := Filters{
		Equals: make(map[string]any),
		Like:   make(map[string]string),
		Range:  make(map[string]DateRange),
	}

	fieldInfos := buildFieldInfoMap(example)

	for key, values := range values {
		if len(values) == 0 {
			continue
		}

		if _, ok := reservedQueryParams[key]; ok {
			continue
		}

		value := strings.TrimSpace(values[0])
		if value == "" {
			continue
		}

		if strings.HasSuffix(key, "_from") || strings.HasSuffix(key, "_to") {
			fieldName, isFrom := splitRangeKey(key)
			info, ok := fieldInfos[fieldName]
			if !ok || !info.IsTime {
				continue
			}

			parsed, err := parseDateTime(value)
			if err != nil {
				return Filters{}, &apperror.AppError{
					HTTPStatus: http.StatusBadRequest,
					Code:       apperror.CodeInvalidParam,
					Message:    "invalid query parameter",
					Detail: []apperror.FieldError{{
						Field:   key,
						Message: fmt.Sprintf("%s must be a valid date/time", key),
					}},
				}
			}

			current := filters.Range[fieldName]
			if isFrom {
				current.From = &parsed
			} else {
				current.To = &parsed
			}
			filters.Range[fieldName] = current
			continue
		}

		info, ok := fieldInfos[key]
		if !ok {
			continue
		}

		if info.IsString {
			filters.Like[key] = value
			continue
		}

		if info.IsBool {
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return Filters{}, &apperror.AppError{
					HTTPStatus: http.StatusBadRequest,
					Code:       apperror.CodeInvalidParam,
					Message:    "invalid query parameter",
					Detail: []apperror.FieldError{{
						Field:   key,
						Message: fmt.Sprintf("%s must be a boolean", key),
					}},
				}
			}
			filters.Equals[key] = parsed
			continue
		}

		if info.IsInt {
			parsed, err := strconv.Atoi(value)
			if err != nil {
				return Filters{}, &apperror.AppError{
					HTTPStatus: http.StatusBadRequest,
					Code:       apperror.CodeInvalidParam,
					Message:    "invalid query parameter",
					Detail: []apperror.FieldError{{
						Field:   key,
						Message: fmt.Sprintf("%s must be an integer", key),
					}},
				}
			}
			filters.Equals[key] = parsed
			continue
		}

		filters.Equals[key] = value
	}

	return filters, nil
}

func ApplyFilters(db *gorm.DB, filters Filters) *gorm.DB {
	for field, value := range filters.Equals {
		db = db.Where(fmt.Sprintf("%s = ?", field), value)
	}

	for field, value := range filters.Like {
		db = db.Where(fmt.Sprintf("%s ILIKE ?", field), "%"+value+"%")
	}

	for field, rng := range filters.Range {
		if rng.From != nil {
			db = db.Where(fmt.Sprintf("%s >= ?", field), *rng.From)
		}
		if rng.To != nil {
			db = db.Where(fmt.Sprintf("%s <= ?", field), *rng.To)
		}
	}

	return db
}

type fieldInfo struct {
	IsString bool
	IsTime   bool
	IsBool   bool
	IsInt    bool
}

func buildFieldInfoMap(example interface{}) map[string]fieldInfo {
	out := make(map[string]fieldInfo)
	rv := reflect.ValueOf(example)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	typeOf := rv.Type()
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		if field.PkgPath != "" {
			continue
		}

		column := getColumnName(field)
		if column == "" {
			continue
		}

		info := fieldInfo{}
		switch field.Type.Kind() {
		case reflect.String:
			info.IsString = true
		case reflect.Bool:
			info.IsBool = true
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			info.IsInt = true
		case reflect.Struct:
			if field.Type == reflect.TypeOf(time.Time{}) {
				info.IsTime = true
			}
		case reflect.Pointer:
			if field.Type.Elem() == reflect.TypeOf(time.Time{}) {
				info.IsTime = true
			}
		}

		if info.IsString || info.IsBool || info.IsInt || info.IsTime {
			out[column] = info
			jsonName := getJSONName(field)
			if jsonName != "" && jsonName != column {
				out[jsonName] = info
			}
		}
	}

	return out
}

func getColumnName(field reflect.StructField) string {
	tag := field.Tag.Get("gorm")
	for _, part := range strings.Split(tag, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}

	jsonName := getJSONName(field)
	if jsonName != "" {
		return jsonName
	}

	return toSnakeCase(field.Name)
}

func getJSONName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return ""
	}
	parts := strings.Split(tag, ",")
	if parts[0] == "-" || parts[0] == "" {
		return ""
	}
	return parts[0]
}

func splitRangeKey(key string) (string, bool) {
	if strings.HasSuffix(key, "_from") {
		return strings.TrimSuffix(key, "_from"), true
	}
	return strings.TrimSuffix(key, "_to"), false
}

func toSnakeCase(value string) string {
	var out strings.Builder
	for i, r := range value {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out.WriteByte('_')
		}
		out.WriteRune(r)
	}
	return strings.ToLower(out.String())
}

func parseDateTime(raw string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid datetime format")
}
