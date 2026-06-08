package helper

import (
	"fmt"
	"os"
	"strings"
)

func BuildFileURL(path string) string {
	if path == "" {
		return ""
	}

	if strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") {
		return path
	}

	baseURL := os.Getenv("APP_BASE_URL")

	path = strings.TrimPrefix(path, "/")

	return fmt.Sprintf(
		"%s/api/v1/files/TzuChiApp/%s",
		baseURL,
		path,
	)
}
