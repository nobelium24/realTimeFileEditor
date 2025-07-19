package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Document struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Title     string    `gorm:"type:varchar(255)" json:"title"`
	Content   *string   `gorm:"type:text" json:"content"`
	UserID    uuid.UUID `gorm:"type:uuid" json:"userId"`
	CreatedAt time.Time `gorm:"type:timestamp" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"updatedAt"`
}

func (d *Document) BeforeCreate(tx *gorm.DB) error {
	d.ID = uuid.New()
	d.CreatedAt = time.Now().UTC()
	d.UpdatedAt = time.Now().UTC()
	return nil
}
