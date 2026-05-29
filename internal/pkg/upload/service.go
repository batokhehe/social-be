package upload

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
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

	// Generate unique filename
	filename := generateUniqueFilename(file.Filename)

	mountPath := os.Getenv("NAS_MOUNT_PATH")
	if mountPath == "" {
		mountPath = "/mnt/nas"
	}
	if err := ensureMountReady(mountPath); err != nil {
		return "", err
	}

	modulePath, err := safeModulePath(module)
	if err != nil {
		return "", err
	}

	targetDir := filepath.Join(mountPath, modulePath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	targetPath := filepath.Join(targetDir, filename)
	dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create upload file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy uploaded file: %w", err)
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	return fmt.Sprintf("%s/%s", modulePath, filename), nil
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
	baseName := filepath.Base(originalFilename)
	ext := filepath.Ext(baseName)
	name := strings.TrimSuffix(baseName, ext)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d%s", name, timestamp, ext)
}

func ensureMountReady(mountPath string) error {
	info, err := os.Stat(mountPath)
	if err != nil {
		return fmt.Errorf("NAS mount path is not accessible: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("NAS mount path is not a directory: %s", mountPath)
	}
	if os.Getenv("NAS_SKIP_MOUNT_CHECK") == "true" {
		return nil
	}
	if !isMountPoint(mountPath) {
		return fmt.Errorf("NAS mount path is not mounted: %s", mountPath)
	}
	return nil
}

func isMountPoint(mountPath string) bool {
	data, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return false
	}
	cleanMountPath := filepath.Clean(mountPath)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 5 && filepath.Clean(unescapeMountInfoPath(fields[4])) == cleanMountPath {
			return true
		}
	}
	return false
}

func unescapeMountInfoPath(value string) string {
	value = strings.ReplaceAll(value, `\040`, " ")
	value = strings.ReplaceAll(value, `\011`, "\t")
	value = strings.ReplaceAll(value, `\012`, "\n")
	value = strings.ReplaceAll(value, `\134`, `\`)
	return value
}

func safeModulePath(module string) (string, error) {
	cleaned := filepath.Clean(module)
	if cleaned == "." || filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, "..") {
		return "", fmt.Errorf("invalid upload module: %s", module)
	}
	return cleaned, nil
}
