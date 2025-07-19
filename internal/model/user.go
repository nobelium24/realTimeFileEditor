package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	FirstName    *string   `gorm:"type:varchar(255)" json:"firstName"`
	LastName     *string   `gorm:"type:varchar(255)" json:"lastName"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	Password     *string   `gorm:"type:varchar(255)" json:"password"`
	ProfilePhoto *Media    `gorm:"type:jsonb" json:"profilePhoto"`
	CreatedAt    time.Time `gorm:"type:timestamp" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"type:timestamp" json:"updatedAt"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = time.Now().UTC()
	return nil
}
