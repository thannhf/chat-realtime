package postgres

import (
	"context"
	"media-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mediaRepository struct {
	db *gorm.DB 
}

func NewMediaRepository(db *gorm.DB) domain.MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Save(ctx context.Context, media *domain.Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}

func (r *mediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	var media domain.Media
	err := r.db.WithContext(ctx).First(&media, "id = ?", id).Error
	if err != nil {
		return nil, err 
	}
	return &media, nil 
}

func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Media{}, "id = ?", id).Error
}