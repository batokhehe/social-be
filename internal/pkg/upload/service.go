package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) UploadFile(ctx context.Context, file *multipart.FileHeader, module string) (string, error) {
	// Validate file
	if err := validateFile(file); err != nil {
		return "", err
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read file content
	fileContent, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Generate unique filename
	filename := generateUniqueFilename(file.Filename)

	// Prepare multipart form data for NAS API
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add file field
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := fw.Write(fileContent); err != nil {
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	// Add module parameter if needed by NAS API
	if err := w.WriteField("module", module); err != nil {
		return "", fmt.Errorf("failed to add module field: %w", err)
	}

	// Close the writer
	w.Close()

	// Get NAS API URL from environment
	nasAPIURL := os.Getenv("NAS_API_URL")
	if nasAPIURL == "" {
		nasAPIURL = "http://localhost:8081/api/store-image" // Default NAS API URL
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", nasAPIURL, &b)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call NAS API: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("NAS API returned status %d", resp.StatusCode)
	}

	// Parse response
	var nasResponse struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&nasResponse); err != nil {
		return "", fmt.Errorf("failed to parse NAS response: %w", err)
	}

	if nasResponse.URL == "" {
		return "", errors.New("NAS API did not return a valid URL")
	}

	// Return path in format {module}/{filename}
	return fmt.Sprintf("%s/%s", module, filename), nil
}

func validateFile(file *multipart.FileHeader) error {
	const maxFileSize = 5 << 20 // 5MB
	if file.Size > maxFileSize {
		return fmt.Errorf("file size exceeds %d MB limit", maxFileSize/(1<<20))
	}

	// Check file extension
	allowedExt := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	return nil
}

func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	name := strings.TrimSuffix(originalFilename, ext)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d%s", name, timestamp, ext)
}
