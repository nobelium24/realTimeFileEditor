package repositories

import (
	"realTimeEditor/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentMediaRepository struct {
	db *gorm.DB
}

func NewDocumentMediaRepository(db *gorm.DB) *DocumentMediaRepository {
	return &DocumentMediaRepository{
		db: db,
	}
}

// Create new media record
func (r *DocumentMediaRepository) Create(media *model.DocumentMedia) error {
	return r.db.Create(media).Error
}

// Get media by ID
func (r *DocumentMediaRepository) GetOne(id uuid.UUID) (*model.DocumentMedia, error) {
	var media model.DocumentMedia
	err := r.db.Where("id = ?", id).First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// Get media by DocumentID
func (r *DocumentMediaRepository) GetByDocumentID(documentID uuid.UUID) ([]model.DocumentMedia, error) {
	var media []model.DocumentMedia
	err := r.db.Where("document_id = ?", documentID).Find(&media).Error
	if err != nil {
		return nil, err
	}
	return media, nil
}

// Update existing media
func (r *DocumentMediaRepository) Update(id uuid.UUID, updated *model.DocumentMedia) error {
	updated.UpdatedAt = time.Now().UTC()
	return r.db.Model(&model.DocumentMedia{}).
		Where("id = ?", id).
		Updates(updated).Error
}

// Delete media by ID
func (r *DocumentMediaRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.DocumentMedia{}, "id = ?", id).Error
}

// Delete all media by DocumentID
func (r *DocumentMediaRepository) DeleteByDocumentID(documentID uuid.UUID) error {
	return r.db.Where("document_id = ?", documentID).Delete(&model.DocumentMedia{}).Error
}
