package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentAccess struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	CollaboratorId uuid.UUID `gorm:"type:uuid" json:"collaboratorId"`
	DocumentId     uuid.UUID `gorm:"type:uuid" json:"documentId"`
	Role           Role      `gorm:"type:varchar" json:"role"`
	CreatedAt      time.Time `gorm:"type:timestamp" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"type:timestamp" json:"updatedAt"`

	Document Document `gorm:"foreignKey:DocumentId"`
	User     User     `gorm:"foreignKey:CollaboratorId"`
}

func (d *DocumentAccess) BeforeCreate(tx *gorm.DB) error {
	d.ID = uuid.New()
	d.CreatedAt = time.Now().UTC()
	d.UpdatedAt = time.Now().UTC()
	return nil
}
