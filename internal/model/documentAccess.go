package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentAccess struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	CollaboratorId uuid.UUID `gorm:"type:uuid" json:"collaboratorId"`
	Role           Role      `gorm:"type:varchar" json:"role"`
	CreatedAt      time.Time `gorm:"type:timestamp" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"type:timestamp" json:"updatedAt"`
}

func (d *DocumentAccess) BeforeCreate(tx *gorm.DB) {
	d.ID = uuid.New()
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
}
