package repositories

import (
	"errors"
	"fmt"
	"math"
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

func (d *DocumentAccessRepository) GetUserDocumentAccesses(userId uuid.UUID) ([]model.DocumentAccess, error) {
	var documentAccesses []model.DocumentAccess
	if err := d.db.Where("collaborator_id = ?", userId).Find(&documentAccesses).Error; err != nil {
		return nil, fmt.Errorf("error fetching documents: %s", err)
	}
	return documentAccesses, nil
}

func (d *DocumentAccessRepository) GetOne(id uuid.UUID, documentAccess *model.DocumentAccess) error {
	return d.db.Where("id = ?", id).First(&documentAccess).Error
}

func (d *DocumentAccessRepository) GetOneWithDocIdAndCollaboratorId(
	docId uuid.UUID, collaboratorId uuid.UUID, documentAccess *model.DocumentAccess) error {
	return d.db.Where(
		"document_id = ? AND collaborator_id = ?", docId, collaboratorId,
	).First(&documentAccess).Error
}

// func (d *DocumentAccessRepository) GetOneWithCollaboratorId(id uuid.UUID, documentAccess *model.DocumentAccess) error {
// 	return d.db.Where("collaborator_id = ?", id).First(&documentAccess).Error
// }collaborator)id

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

func (d *DocumentAccessRepository) HasEditAccess(userId, docId uuid.UUID) (bool, error) {
	var access model.DocumentAccess
	err := d.db.Where("collaborator_id = ? AND document_id = ? AND role IN ?", userId, docId, []model.Role{model.Edit, model.Creator}).
		First(&access).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (d *DocumentAccessRepository) HasReadAccess(userId, docId uuid.UUID) (bool, error) {
	var access model.DocumentAccess
	err := d.db.Where("collaborator_id = ? AND document_id = ?", userId, docId).First(&access).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (d *DocumentAccessRepository) UpdateWithTransaction(tx *gorm.DB, document *model.DocumentAccess, id uuid.UUID) error {
	if err := tx.Where("id = ?", id).First(document).Error; err != nil {
		return err
	}
	document.UpdatedAt = time.Now().UTC()
	return tx.Model(&model.DocumentAccess{}).
		Where("id = ?", id).Updates(document).Error
}

func (d *DocumentAccessRepository) ExecuteInTransaction(fn func(tx *gorm.DB) error, maxRetries int) error {
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
