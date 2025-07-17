package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"realTimeEditor/internal/model"
	"realTimeEditor/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentController struct {
	DocumentRepository       *repositories.DocumentRepository
	DocumentAccessRepository *repositories.DocumentAccessRepository
}

func NewDocumentController(
	documentRepository *repositories.DocumentRepository,
	documentAccessRepository *repositories.DocumentAccessRepository,
) *DocumentController {
	return &DocumentController{
		DocumentRepository:       documentRepository,
		DocumentAccessRepository: documentAccessRepository,
	}
}

func (d *DocumentController) Create(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	var newDocument model.Document
	if err := c.ShouldBindJSON(&newDocument); err != nil {
		log.Printf("Error binding JSON: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	newDocument.UserID = userDetails.ID

	if err := d.DocumentRepository.Create(&newDocument); err != nil {
		log.Printf("Error creating document: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	newDocumentAccess := model.DocumentAccess{
		CollaboratorId: userDetails.ID,
		Role:           model.Creator,
		DocumentId:     newDocument.ID,
	}

	if err := d.DocumentAccessRepository.Create(&newDocumentAccess); err != nil {
		log.Printf("Error creating document: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Document created successfully"})
}

func (d *DocumentController) GetUserCreatedDocuments(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	documents, err := d.DocumentRepository.GetUserDocuments(userDetails.ID)
	if err != nil {
		log.Printf("Error fetching document: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Documents fetched", "documents": documents})
}

func (d *DocumentController) GetSingleDocument(c *gin.Context) {
	user, exists := c.Get("user")
	documentId := c.Param("historyId")

	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	documentUUID, err := uuid.Parse(documentId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching document"})
		return
	}

	var document model.Document
	if err := d.DocumentRepository.GetOne(documentUUID, &document); err != nil {
		log.Printf("Error fetching document: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	access, err := d.DocumentAccessRepository.HasReadAccess(userDetails.ID, documentUUID)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching document"})
		return
	}

	if !access {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you do not have access to this document"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Document fetched", "document": document})
}

func (d *DocumentController) RevokeAccess(c *gin.Context) {
	user, exists := c.Get("user")
	documentAccessId := c.Param("documentAccessId")

	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	documentAccessUUID, err := uuid.Parse(documentAccessId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document access ID"})
		return
	}

	var documentAccess model.DocumentAccess
	if err := d.DocumentAccessRepository.GetOne(documentAccessUUID, &documentAccess); err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "document access not found"})
		return
	}

	if documentAccess.Role == model.Creator {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot revoke access to a document created by you"})
		return
	}

	var requesterAccess model.DocumentAccess
	err = d.DocumentAccessRepository.GetOneWithDocIdAndCollaboratorId(documentAccess.DocumentId, userDetails.ID, &requesterAccess)
	if err != nil {
		log.Printf("Error checking requester access: %s", err.Error())
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if requesterAccess.Role != model.Creator {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can revoke access"})
		return
	}

	if err := d.DocumentAccessRepository.Delete(&documentAccess, documentAccessUUID); err != nil {
		log.Printf("Error deleting document access: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access revoked successfully"})
}

func (d *DocumentController) DeleteDocument(c *gin.Context) {
	user, exists := c.Get("user")
	documentId := c.Param("documentId")

	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	documentUUID, err := uuid.Parse(documentId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document access ID"})
		return
	}

	var document model.Document
	if err := d.DocumentRepository.GetOne(documentUUID, &document); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "document not found"})
			return
		}
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if userDetails.ID != document.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you do not have access to carry out delete action"})
		return
	}

	if err := d.DocumentRepository.Delete(documentUUID); err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted"})
}

func (d *DocumentController) ModifyAccess(c *gin.Context) {
	user, exists := c.Get("user")
	documentAccessId := c.Param("documentAccessId")
	newRole := c.Param("newRole")

	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	documentAccessUUID, err := uuid.Parse(documentAccessId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document access ID"})
		return
	}

	var documentAccess model.DocumentAccess
	if err := d.DocumentAccessRepository.GetOne(documentAccessUUID, &documentAccess); err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "document access not found"})
		return
	}

	if documentAccess.Role == model.Creator && documentAccess.CollaboratorId == userDetails.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot modify your role as document creator"})
		return
	}

	var requesterAccess model.DocumentAccess
	err = d.DocumentAccessRepository.GetOneWithDocIdAndCollaboratorId(documentAccess.DocumentId, userDetails.ID, &requesterAccess)
	if err != nil {
		log.Printf("Error checking requester access: %s", err.Error())
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if requesterAccess.Role != model.Creator {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can modify role"})
		return
	}

	if newRole == string(model.Creator) {
		c.JSON(http.StatusForbidden, gin.H{"error": "a document can only have one creator"})
		return
	}

	validRoles := map[string]bool{
		string(model.Edit): true,
		string(model.Read): true,
	}

	if !validRoles[newRole] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	documentAccess.Role = model.Role(newRole)

	if err := d.DocumentAccessRepository.Update(&documentAccess, documentAccessUUID); err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
}

func (d *DocumentController) FetchAllDocuments(c *gin.Context) {
	user, exists := c.Get("user")

	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	documents, err := d.DocumentAccessRepository.GetUserDocumentAccesses(userDetails.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"message": "You do not have documents yet"})
			return
		}
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Documents fetched", "documents": documents})
}

func (d *DocumentController) FetchCollaborators(c *gin.Context) {
	user, exists := c.Get("user")
	documentId := c.Param("documentId")

	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	documentUUID, err := uuid.Parse(documentId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document access ID"})
		return
	}

	var docAccess model.DocumentAccess
	if err := d.DocumentAccessRepository.GetOneWithDocIdAndCollaboratorId(documentUUID, userDetails.ID, &docAccess); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You don't have necessary permission to view this"})
			return
		}
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	collaborators, err := d.DocumentAccessRepository.GetDocumentAccesses(documentUUID)
	if err := d.DocumentAccessRepository.GetOneWithDocIdAndCollaboratorId(documentUUID, userDetails.ID, &docAccess); err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collaborators fetched", "collaborators": collaborators})
}

func (d *DocumentController) TransferOwnership(c *gin.Context) {
	user, exists := c.Get("user")
	documentId := c.Param("documentId")
	recipientId := c.Param("recipientId")

	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	documentUUID, err := uuid.Parse(documentId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document access ID"})
		return
	}

	recipientUUID, err := uuid.Parse(recipientId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document access ID"})
		return
	}

	var document model.Document
	if err := d.DocumentRepository.GetOne(documentUUID, &document); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if userDetails.ID != document.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you do not have access to carry out this action"})
		return
	}

	var accessOne model.DocumentAccess
	if err := d.DocumentAccessRepository.GetOneWithDocIdAndCollaboratorId(documentUUID, userDetails.ID, &accessOne); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "document access for creator not found"})
			return
		}
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var accessTwo model.DocumentAccess
	if err := d.DocumentAccessRepository.GetOneWithDocIdAndCollaboratorId(documentUUID, recipientUUID, &accessTwo); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "document access for collaborator not found"})
			return
		}
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	document.UserID = recipientUUID
	accessOne.Role = model.Edit
	accessTwo.Role = model.Creator

	err = d.DocumentRepository.ExecuteInTransaction(func(tx *gorm.DB) error {
		if err := d.DocumentRepository.UpdateWithTransaction(tx, &document, documentUUID); err != nil {
			log.Printf("Error: %s", err)
			return fmt.Errorf("failed to update document details: %w", err)
		}

		if err := d.DocumentAccessRepository.UpdateWithTransaction(tx, &accessOne, accessOne.ID); err != nil {
			log.Printf("Error: %s", err)
			return fmt.Errorf("failed to update document details: %w", err)
		}

		if err := d.DocumentAccessRepository.UpdateWithTransaction(tx, &accessTwo, accessTwo.ID); err != nil {
			log.Printf("Error: %s", err)
			return fmt.Errorf("failed to update document details: %w", err)
		}

		return nil
	}, 3)

	if err != nil {
		log.Printf("Transaction failed: %s", err)
		if originalErr := errors.Unwrap(err); originalErr != nil {
			err = originalErr
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error transferring ownership"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document ownership transferred"})
}
