package repositories

import (
	"realTimeEditor/internal/model"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (u *UserRepository) Create(member *model.User) (*model.User, error) {
	err := u.db.Create(member).Error
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (u *UserRepository) Update(member *model.User, id uuid.UUID) error {
	if err := u.db.Where("id = ?", id).First(member).Error; err != nil {
		return err
	}
	return u.db.Save(member).Error
}

func (u *UserRepository) GetById(member *model.User, id uuid.UUID) error {
	return u.db.Where("id = ?", id).First(&member).Error
}

func (u *UserRepository) GetByEmail(member *model.User, email string) error {
	return u.db.Where("email = ?", email).First(&member).Error
}
