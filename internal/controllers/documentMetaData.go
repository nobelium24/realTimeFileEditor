package controllers

import (
	"errors"
	"log"
	"net/http"
	"realTimeEditor/internal/model"
	"realTimeEditor/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentMetadataController struct {
	DocumentRepository         *repositories.DocumentRepository
	DocumentMetaDataRepository *repositories.DocumentMetaDataRepository
}

func NewDocumentMetaDataController(
	documentRepository *repositories.DocumentRepository,
	documentMetaDataRepository *repositories.DocumentMetaDataRepository,
) *DocumentMetadataController {
	return &DocumentMetadataController{
		DocumentRepository:         documentRepository,
		DocumentMetaDataRepository: documentMetaDataRepository,
	}
}

func (d *DocumentMetadataController) Create(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	_, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	var newDocumentMetaData model.DocumentMetadata
	if err := c.ShouldBindJSON(&newDocumentMetaData); err != nil {
		log.Printf("Error binding JSON: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := d.DocumentMetaDataRepository.Create(&newDocumentMetaData); err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Document metadata created successfully"})
}

func (d *DocumentMetadataController) GetDocumentMetadata(c *gin.Context) {
	user, exists := c.Get("user")
	documentId := c.Param("documentId")

	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	_, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	documentUUID, err := uuid.Parse(documentId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var documentMetaData model.DocumentMetadata
	if err := d.DocumentMetaDataRepository.GetOneByDocId(documentUUID, &documentMetaData); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "documentMetaData not found"})
			return
		}
		log.Printf("Error fetching documentMetaData: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document metadata fetched", "metadata": documentMetaData})
}

func (d *DocumentMetadataController) Update(c *gin.Context) {
	user, exists := c.Get("user")
	documentMetadataId := c.Param("documentMetadataId")

	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	_, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	documentMetadataUUID, err := uuid.Parse(documentMetadataId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var documentMetaData model.DocumentMetadata
	if err := c.ShouldBindJSON(&documentMetaData); err != nil {
		log.Printf("Error binding JSON: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	var existingMetadata model.DocumentMetadata
	if err := d.DocumentMetaDataRepository.GetOne(documentMetadataUUID, &existingMetadata); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "documentMetaData not found"})
			return
		}
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if err := d.DocumentMetaDataRepository.Update(&documentMetaData, documentMetadataUUID); err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document metadata updated"})
}

func (d *DocumentMetadataController) Delete(c *gin.Context) {
	user, exists := c.Get("user")
	documentMetadataId := c.Param("documentMetadataId")

	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	_, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	documentMetadataUUID, err := uuid.Parse(documentMetadataId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var documentMetaData model.DocumentMetadata
	if err := d.DocumentMetaDataRepository.Delete(&documentMetaData, documentMetadataUUID); err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document metadata deleted"})
}
