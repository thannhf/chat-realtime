package usecase

import (
	"context"
	"fmt"
	"io"
	"media-service/config"
	"media-service/internal/domain"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type MediaUsecase interface {
	UploadFile(ctx context.Context, uploaderID uuid.UUID, filename string, reader io.Reader, size int64) (*domain.Media, error)
	GetMediaByID(ctx context.Context, id uuid.UUID) (*domain.Media, error)
}

type mediaUsecase struct {
	mediaRepo domain.MediaRepository
	storageEngine domain.StorageEngine
	cfg *config.Config
}

func NewMediaUsecase(repo domain.MediaRepository, engine domain.StorageEngine, cfg *config.Config) MediaUsecase {
	return &mediaUsecase{
		mediaRepo: repo,
		storageEngine: engine,
		cfg: cfg,
	}
}

func (u *mediaUsecase) UploadFile(ctx context.Context, uploaderID uuid.UUID, filename string, reader io.Reader, size int64) (*domain.Media, error) {
	if size > u.cfg.Storage.MaxFileSize {
		return nil, fmt.Errorf("dung lượng file (%d bytes) vượt quá giới hạn cho phép (%d bytes)", size, u.cfg.Storage.MaxFileSize)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	var mediaType domain.MediaType

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		mediaType = domain.TypeImage
	case ".mp4", ".mov", ".avi", ".mkv":
		mediaType = domain.TypeVideo
	default:
		mediaType = domain.TypeFile
	}

	fileURL, err := u.storageEngine.Upload(ctx, filename, reader, size)
	if err != nil {
		return nil, fmt.Errorf("Lỗi trong quá trình ghi file vật lý: %w", err)
	}

	mediaID :=uuid.New()
	newMedia := &domain.Media{
		ID: mediaID,
		FileName: filename,
		FileURL: fileURL,
		FileType: mediaType,
		FileSize: size,
		UploaderID: uploaderID,
	}

	if err := u.mediaRepo.Save(ctx, newMedia); err != nil {
		_ = u.storageEngine.Delete(ctx, fileURL)
		return nil, fmt.Errorf("lỗi lưu trữ metadata vào DB: %w", err)
	}

	return newMedia, nil 
}

func (u *mediaUsecase) GetMediaByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	return u.mediaRepo.GetByID(ctx, id)
}