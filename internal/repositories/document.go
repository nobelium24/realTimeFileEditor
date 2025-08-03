package repositories

import (
	"fmt"
	"math"
	"realTimeEditor/internal/model"

	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{
		db: db,
	}
}

func (d *DocumentRepository) Create(document *model.Document) error {
	return d.db.Create(document).Error
}

func (d *DocumentRepository) CreateWithTransaction(tx *gorm.DB, document *model.Document) error {
	return tx.Create(document).Error
}

func (d *DocumentRepository) GetUserDocuments(userId uuid.UUID) ([]model.Document, error) {
	var documents []model.Document
	if err := d.db.Where("user_id = ?", userId).Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("error fetching documents: %s", err)
	}
	return documents, nil
}

func (d *DocumentRepository) ToggleVisibility(id uuid.UUID) error {
	var existingDoc model.Document

	if err := d.db.First(&existingDoc, "id = ?", id).Error; err != nil {
		return err
	}
	existingDoc.PublicVisibility = !existingDoc.PublicVisibility
	existingDoc.UpdatedAt = time.Now().UTC()
	return d.db.Save(&existingDoc).Error
}

func (d *DocumentRepository) Update(updatedDoc *model.Document, id uuid.UUID) error {
	var existingDoc model.Document

	if err := d.db.First(&existingDoc, "id = ?", id).Error; err != nil {
		return err
	}

	existingDoc.Title = updatedDoc.Title
	existingDoc.Content = updatedDoc.Content
	existingDoc.UpdatedAt = time.Now().UTC()

	return d.db.Save(&existingDoc).Error
}

func (d *DocumentRepository) UpdateWithTransaction(tx *gorm.DB, document *model.Document, id uuid.UUID) error {
	if err := tx.Where("id = ?", id).First(document).Error; err != nil {
		return err
	}
	document.UpdatedAt = time.Now().UTC().UTC()
	return tx.Model(&model.Document{}).
		Where("id = ?", id).Updates(document).Error
}

func (d *DocumentRepository) Delete(id uuid.UUID) error {
	return d.db.Delete(&model.Document{}, "id = ?", id).Error
}

func (d *DocumentRepository) DeleteWithTransaction(tx *gorm.DB, id uuid.UUID) error {
	return tx.Delete(&model.Document{}, "id = ?", id).Error
}

func (d *DocumentRepository) GetOne(id uuid.UUID, document *model.Document) error {
	return d.db.Where("id = ?", id).First(&document).Error
}

func (d *DocumentRepository) GetOneWithTransaction(tx *gorm.DB, id uuid.UUID, document *model.Document) error {
	return tx.Where("id = ?", id).First(&document).Error
}

func (d *DocumentRepository) ExecuteInTransaction(fn func(tx *gorm.DB) error, maxRetries int) error {
	var lastErr error
	for i := range maxRetries {
		err := d.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ").Error; err != nil {
				return err
			}
			return fn(tx)
		})

		if err == nil {
			return nil
		}

		time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Millisecond * 100)
		lastErr = err
	}

	return fmt.Errorf("transaction failed after %d retries: %w", maxRetries, lastErr)
}
