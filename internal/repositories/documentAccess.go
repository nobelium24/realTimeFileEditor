package repositories

import (
	"errors"
	"fmt"
	"realTimeEditor/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentAccessRepository struct {
	db *gorm.DB
}

func NewDocumentAccessRepository(db *gorm.DB) *DocumentAccessRepository {
	return &DocumentAccessRepository{
		db: db,
	}
}

func (d *DocumentAccessRepository) Create(documentAccess *model.DocumentAccess) error {
	return d.db.Create(documentAccess).Error
}

func (d *DocumentAccessRepository) GetDocumentAccesses(documentId uuid.UUID) ([]model.DocumentAccess, error) {
	var documentAccesses []model.DocumentAccess

	err := d.db.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "first_name", "last_name")
		}).
		Where("document_id = ?", documentId).
		Find(&documentAccesses).Error

	if err != nil {
		return nil, fmt.Errorf("error fetching document accesses: %w", err)
	}

	return documentAccesses, nil
}

func (d *DocumentAccessRepository) GetOne(id uuid.UUID, documentAccess *model.DocumentAccess) error {
	return d.db.Where("id = ?", id).First(&documentAccess).Error
}

func (d *DocumentAccessRepository) Update(documentAccess *model.DocumentAccess, id uuid.UUID) error {
	if err := d.db.Where("id = ?", id).Error; err != nil {
		return err
	}
	documentAccess.UpdatedAt = time.Now().UTC()
	return d.db.Model(&model.DocumentAccess{}).
		Where("id = ?", id).Updates(documentAccess).Error
}

func (d *DocumentAccessRepository) Delete(documentAccess *model.DocumentAccess, id uuid.UUID) error {
	return d.db.Delete(documentAccess, "id = ?", id).Error
}

func (d *DocumentAccessRepository) HasEditAccess(userId, docId string) (bool, error) {
	var access model.DocumentAccess
	err := d.db.Where("collaborator_id = ? AND document_id = ? AND role = ?", userId, docId, model.Edit).First(&access).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
