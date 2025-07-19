package repositories

import (
	"fmt"
	"realTimeEditor/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentMetaDataRepository struct {
	db *gorm.DB
}

func NewDocumentMetaDataRepository(db *gorm.DB) *DocumentMetaDataRepository {
	return &DocumentMetaDataRepository{
		db: db,
	}
}

func (d *DocumentMetaDataRepository) Create(metaData *model.DocumentMetadata) error {
	return d.db.Create(metaData).Error
}

func (d *DocumentMetaDataRepository) GetAll() ([]model.DocumentMetadata, error) {
	var metaData []model.DocumentMetadata
	if err := d.db.Find(&metaData).Error; err != nil {
		return nil, fmt.Errorf("error fetching metaData: %s", err)
	}
	return metaData, nil
}

func (d *DocumentMetaDataRepository) GetOne(id uuid.UUID, metaData *model.DocumentMetadata) error {
	return d.db.Where("id = ?", id).First(metaData).Error
}

func (d *DocumentMetaDataRepository) GetOneByDocId(documentId uuid.UUID, metaData *model.DocumentMetadata) error {
	return d.db.Where("document_id = ?", documentId).First(metaData).Error
}

func (d *DocumentMetaDataRepository) Update(metaData *model.DocumentMetadata, id uuid.UUID) error {
	if err := d.db.Where("id = ?", id).Error; err != nil {
		return err
	}
	metaData.UpdatedAt = time.Now().UTC()
	return d.db.Model(&model.DocumentMetadata{}).
		Where("id = ?", id).Updates(metaData).Error
}

func (d *DocumentMetaDataRepository) Delete(metaData *model.DocumentMetadata, id uuid.UUID) error {
	return d.db.Delete(metaData, "id = ?", id).Error
}
