package storage

import (
	"context"
	"fmt"
	"io"
	"media-service/config"
	"media-service/internal/domain"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type localStorageEngine struct {
	cfg *config.Config
}

func NewLocalStorageEngine(cfg *config.Config) domain.StorageEngine {
	return &localStorageEngine{cfg: cfg}
}

func (s *localStorageEngine) Upload(ctx context.Context, filename string, reader io.Reader, size int64) (string, error) {
	ext := filepath.Ext(filename)
	cleanName := strings.TrimSuffix(filepath.Base(filename), ext)

	uniqueName := fmt.Sprintf("%s_%s%s", time.Now().Format("20060102"), uuid.New().String(), ext)
	_ = cleanName

	targetPath := filepath.Join(s.cfg.Storage.UploadDir, uniqueName)

	out, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("Không thể tạo file vật lý trên server: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	if err != nil {
		return "", fmt.Errorf("Lỗi trong quá trình stream ghi file: %w", err)
	}

	fileURL := fmt.Sprintf("/uploads/%s", uniqueName)
	return fileURL, nil 
}

func (s *localStorageEngine) Delete(ctx context.Context, fileURL string) error {
	filename := filepath.Base(fileURL)
	targetPath := filepath.Join(s.cfg.Storage.UploadDir, filename)

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Không thể xóa file vật lý: %w", err)
	}

	return nil
}