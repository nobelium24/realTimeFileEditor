package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentMedia struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DocumentID uuid.UUID `gorm:"type:uuid;not null" json:"document_id"`
	PublicID   string    `gorm:"type:text" json:"public_id"`
	SecureURL  string    `gorm:"type:text" json:"secure_url"`
	Format     string    `gorm:"type:varchar(20)" json:"format"` // e.g., "pdf", "docx", etc.

	CreatedAt time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"updated_at"`
}

func (m *DocumentMedia) BeforeCreate(tx *gorm.DB) error {
	m.ID = uuid.New()
	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *DocumentMedia) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now().UTC()
	return nil
}
