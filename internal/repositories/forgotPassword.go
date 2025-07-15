package repositories

import (
	"fmt"
	"realTimeEditor/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ForgotPasswordRepository struct {
	db *gorm.DB
}

func NewForgotPasswordRepository(db *gorm.DB) *ForgotPasswordRepository {
	return &ForgotPasswordRepository{
		db: db,
	}
}

func (f *ForgotPasswordRepository) Create(forgotPassword *model.ForgotPassword) error {
	return f.db.Create(forgotPassword).Error
}

func (f *ForgotPasswordRepository) GetAll() ([]model.ForgotPassword, error) {
	var forgotPasswords []model.ForgotPassword
	if err := f.db.Find(&forgotPasswords).Error; err != nil {
		return nil, fmt.Errorf("error fetching forgotPasswords: %s", err)
	}
	return forgotPasswords, nil
}

func (f *ForgotPasswordRepository) GetOne(id uuid.UUID, forgotPassword *model.ForgotPassword) error {
	return f.db.Where("id = ?", id).First(forgotPassword).Error
}

func (f *ForgotPasswordRepository) GetOneByCode(resetCode string, forgotPassword *model.ForgotPassword) error {
	return f.db.Where("reset_code = ?", resetCode).First(forgotPassword).Error
}

func (f *ForgotPasswordRepository) Update(forgotPassword *model.ForgotPassword, id uuid.UUID) error {
	if err := f.db.Where("id = ?", id).Error; err != nil {
		return err
	}
	forgotPassword.UpdatedAt = time.Now().UTC()
	return f.db.Model(&model.ForgotPassword{}).
		Where("id = ?", id).Updates(forgotPassword).Error
}

func (f *ForgotPasswordRepository) Delete(forgotPassword *model.ForgotPassword, id uuid.UUID) error {
	return f.db.Delete(forgotPassword, "id = ?", id).Error
}
