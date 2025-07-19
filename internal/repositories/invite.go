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

func (i *InviteRepository) Create(invite *model.Invite) error {
	return i.db.Create(invite).Error
}

func (i *InviteRepository) GetAll() ([]model.Invite, error) {
	var invites []model.Invite
	if err := i.db.Find(&invites).Error; err != nil {
		return nil, fmt.Errorf("error fetching invites: %s", err)
	}
	return invites, nil
}

func (i *InviteRepository) GetOne(id uuid.UUID, invite *model.Invite) error {
	return i.db.Where("id = ?", id).First(invite).Error
}

func (i *InviteRepository) GetOneByToken(token string, invite *model.Invite) error {
	return i.db.Where("token = ?", token).First(invite).Error
}

func (i *InviteRepository) Update(invite *model.Invite, id uuid.UUID) error {
	if err := i.db.Where("id = ?", id).Error; err != nil {
		return err
	}
	invite.UpdatedAt = time.Now().UTC().UTC()
	return i.db.Model(&model.Invite{}).
		Where("id = ?", id).Updates(invite).Error
}

func (i *InviteRepository) GetOneByEmailAndDocId(invite *model.Invite, email string, docId uuid.UUID) error {
	return i.db.Where("email = ? AND document_id = ?", email, docId).First(invite).Error
}

func (i *InviteRepository) Delete(invite *model.Invite, id uuid.UUID) error {
	return i.db.Delete(invite, "id = ?", id).Error
}
