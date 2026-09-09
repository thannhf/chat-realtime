package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	TypeImage MediaType = "image"
	TypeVideo MediaType = "video"
	TypeFile  MediaType = "file"
)

type Media struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;"`
	FileName   string    `json:"file_name" gorm:"type:varchar(255);not null"`
	FileURL    string    `json:"file_url" gorm:"type:text;not null"`
	FileType   MediaType `json:"file_type" gorm:"type:varchar(50);not null"`
	FileSize   int64     `json:"file_size" gorm:"type:bigint;not null"`
	UploaderID uuid.UUID `json:"uploader_id" gorm:"type:uuid;not null;index:idx_media_uploader"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:current_timestamp"`
}

type MediaVariant struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MediaID     uuid.UUID `gorm:"type:uuid;not null;index:idx_media_variants_parent" json:"media_id"`
	VariantType string    `gorm:"type:varchar(50);not null" json:"variant_type"`
	FileURL     string    `gorm:"type:text;not null" json:"file_url"`
	FileSize    int64     `gorm:"type:bigint;not null" json:"file_size"`
	Width       int       `gorm:"type:int" json:"width,omitempty"`
	Height      int       `gorm:"type:int" json:"height,omitempty"`
	CreatedAt   time.Time `gorm:"type:timestamp with time zone;default:current_timestamp" json:"created_at"`
}

type MediaRepository interface {
	Save(ctx context.Context, media *Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*Media, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type StorageEngine interface {
	Upload(ctx context.Context, filename string, reader io.Reader, size int64) (string, error)
	Delete(ctx context.Context, fileURL string) error
}
