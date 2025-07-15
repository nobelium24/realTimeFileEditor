package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ForgotPassword struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"type:varchar(255)" json:"email"`
	ResetCode string    `gorm:"type:varchar(255)" json:"resetCode"`
	CreatedAt time.Time `gorm:"type:timestamp" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"updatedAt"`
}

func (forgotPassword *ForgotPassword) BeforeCreate(tx *gorm.DB) error {
	forgotPassword.ID = uuid.New()
	forgotPassword.CreatedAt = time.Now().UTC()
	forgotPassword.UpdatedAt = time.Now().UTC()
	return nil
}
