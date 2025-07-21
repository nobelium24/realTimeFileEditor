package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentMetadata struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	DocumentID uuid.UUID `gorm:"type:uuid" json:"documentId"`
	Version    int       `gorm:"type:int" json:"version"`
	// Cursor     *CursorPosition `json:"cursor,omitempty"`

	Metadata *Metadata `gorm:"type:jsonb" json:"metadata,omitempty"`

	CreatedAt time.Time `gorm:"type:timestamp" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"updatedAt"`

	Document Document `gorm:"foreignKey:DocumentID"`
}

type Metadata struct {
	Font         string  `gorm:"type:varchar" json:"font,omitempty"`
	FontSize     float64 `gorm:"type:decimal(2,1)" json:"fontSize,omitempty"`
	LineSpacing  float64 `gorm:"type:decimal(2,1)" json:"lineSpacing,omitempty"`
	MarginTop    float64 `gorm:"type:decimal(2,1)" json:"marginTop,omitempty"`
	MarginLeft   float64 `gorm:"type:decimal(2,1)" json:"marginLeft,omitempty"`
	MarginRight  float64 `gorm:"type:decimal(2,1)" json:"marginRight,omitempty"`
	MarginBottom float64 `gorm:"type:decimal(2,1)" json:"marginBottom,omitempty"`
}

func (d *DocumentMetadata) BeforeCreate(tx *gorm.DB) error {
	d.ID = uuid.New()
	d.CreatedAt = time.Now().UTC()
	d.UpdatedAt = time.Now().UTC()
	return nil
}
