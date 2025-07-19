package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Invite struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey"`
	CollaboratorId *uuid.UUID   `gorm:"type:uuid;index;default:null" json:"collaboratorId"`
	Email          *string      `gorm:"type:varchar(255);index;not null" json:"email"`
	DocumentId     uuid.UUID    `gorm:"type:uuid;not null;index" json:"documentId"`
	Role           Role         `gorm:"type:varchar(20);not null" json:"role"`
	Token          string       `gorm:"type:varchar(64);not null;uniqueIndex" json:"token"`
	Status         InviteStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	InviterId      uuid.UUID    `gorm:"type:uuid;index;default:null" json:"inviterId"`
	CreatedAt      time.Time    `gorm:"type:timestamp" json:"createdAt"`
	UpdatedAt      time.Time    `gorm:"type:timestamp" json:"updatedAt"`
}

func (i *Invite) BeforeCreate(tx *gorm.DB) error {
	i.ID = uuid.New()
	i.CreatedAt = time.Now().UTC()
	i.UpdatedAt = time.Now().UTC()
	return nil
}
