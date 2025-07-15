package repositories

import (
	"realTimeEditor/internal/model"

	"github.com/google/uuid"
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

func (u *UserRepository) Create(user *model.User) (*model.User, error) {
	err := u.db.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserRepository) Update(user *model.User, id uuid.UUID) error {
	if err := u.db.Where("id = ?", id).First(user).Error; err != nil {
		return err
	}
	return u.db.Save(user).Error
}

func (u *UserRepository) GetById(user *model.User, id uuid.UUID) error {
	return u.db.Where("id = ?", id).First(&user).Error
}

func (u *UserRepository) GetByEmail(user *model.User, email string) error {
	return u.db.Where("email = ?", email).First(&user).Error
}
