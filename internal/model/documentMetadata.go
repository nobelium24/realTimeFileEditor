package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentMetadata struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	DocumentID string    `json:"documentId"`
	Version    int       `json:"version"`
	// Cursor     *CursorPosition `json:"cursor,omitempty"`

	Metadata *Metadata `json:"metadata,omitempty"`

	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now()" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null;default:now()" json:"updatedAt"`
}

type Metadata struct {
	Font         string  `json:"font,omitempty"`        // e.g., "Arial", "Times New Roman"
	FontSize     float64 `json:"fontSize,omitempty"`    // e.g., 12.0
	LineSpacing  float64 `json:"lineSpacing,omitempty"` // e.g., 1.15
	MarginTop    float64 `json:"marginTop,omitempty"`
	MarginLeft   float64 `json:"marginLeft,omitempty"`
	MarginRight  float64 `json:"marginRight,omitempty"`
	MarginBottom float64 `json:"marginBottom,omitempty"`
}

func (d *DocumentMetadata) BeforeCreate(tx *gorm.DB) {
	d.ID = uuid.New()
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
}
