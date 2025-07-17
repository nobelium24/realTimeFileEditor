package repositories

import (
	"fmt"
	"realTimeEditor/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InviteRepository struct {
	db *gorm.DB
}

func NewInviteRepository(db *gorm.DB) *InviteRepository {
	return &InviteRepository{
		db: db,
	}
}

func (f *InviteRepository) Create(invite *model.Invite) error {
	return f.db.Create(invite).Error
}

func (f *InviteRepository) GetAll() ([]model.Invite, error) {
	var invites []model.Invite
	if err := f.db.Find(&invites).Error; err != nil {
		return nil, fmt.Errorf("error fetching invites: %s", err)
	}
	return invites, nil
}

func (f *InviteRepository) GetOne(id uuid.UUID, invite *model.Invite) error {
	return f.db.Where("id = ?", id).First(invite).Error
}

func (f *InviteRepository) GetOneByToken(token string, invite *model.Invite) error {
	return f.db.Where("token = ?", token).First(invite).Error
}

func (f *InviteRepository) Update(invite *model.Invite, id uuid.UUID) error {
	if err := f.db.Where("id = ?", id).Error; err != nil {
		return err
	}
	invite.UpdatedAt = time.Now().UTC()
	return f.db.Model(&model.Invite{}).
		Where("id = ?", id).Updates(invite).Error
}

func (f *InviteRepository) Delete(invite *model.Invite, id uuid.UUID) error {
	return f.db.Delete(invite, "id = ?", id).Error
}
