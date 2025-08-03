package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"realTimeEditor/internal/handlers"
	"realTimeEditor/internal/model"
	"realTimeEditor/internal/repositories"
	"realTimeEditor/pkg/constants"
	"realTimeEditor/pkg/jwt"
	"realTimeEditor/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentController struct {
	DocumentRepository         *repositories.DocumentRepository
	DocumentAccessRepository   *repositories.DocumentAccessRepository
	InviteRepository           *repositories.InviteRepository
	UserRepository             *repositories.UserRepository
	DocumentMetadataRepository *repositories.DocumentMetaDataRepository
	DocumentMediaRepository    *repositories.DocumentMediaRepository
}

func NewDocumentController(
	documentRepository *repositories.DocumentRepository,
	documentAccessRepository *repositories.DocumentAccessRepository,
	inviteRepository *repositories.InviteRepository,
	userRepository *repositories.UserRepository,
	documentMetadataRepository *repositories.DocumentMetaDataRepository,
	documentMediaRepository *repositories.DocumentMediaRepository,
) *DocumentController {
	return &DocumentController{
		DocumentRepository:         documentRepository,
		DocumentAccessRepository:   documentAccessRepository,
		InviteRepository:           inviteRepository,
		UserRepository:             userRepository,
		DocumentMetadataRepository: documentMetadataRepository,
		DocumentMediaRepository:    documentMediaRepository,
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

	documentMetaData := model.DocumentMetadata{
		DocumentID: newDocument.ID,
		Version:    1,
	}
	if err := d.DocumentMetadataRepository.Create(&documentMetaData); err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var document model.Document
	if err := d.DocumentRepository.GetOne(documentUUID, &document); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "document not found"})
			return
		}
		log.Printf("Error fetching document: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if !document.PublicVisibility {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var documentAccess model.DocumentAccess
	if err := d.DocumentAccessRepository.GetOne(documentAccessUUID, &documentAccess); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "document access not found"})
			return
		}
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "document access not found"})
			return
		}
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
	documentId := c.Param("id")

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var documentAccess model.DocumentAccess
	if err := d.DocumentAccessRepository.GetOne(documentAccessUUID, &documentAccess); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "document access not found"})
			return
		}
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "document access not found"})
			return
		}
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
	documentId := c.Param("id")

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "document access not found"})
			return
		}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	recipientUUID, err := uuid.Parse(recipientId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	if userDetails.ID == recipientUUID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you already own this document"})
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

	var creatorAccess model.DocumentAccess
	if err := d.DocumentAccessRepository.GetOneWithDocIdAndCollaboratorId(documentUUID, userDetails.ID, &creatorAccess); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "document access for creator not found"})
			return
		}
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var recipientAccess model.DocumentAccess
	if err := d.DocumentAccessRepository.GetOneWithDocIdAndCollaboratorId(documentUUID, recipientUUID, &recipientAccess); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "document access for collaborator not found"})
			return
		}
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	document.UserID = recipientUUID
	creatorAccess.Role = model.Edit
	recipientAccess.Role = model.Creator

	err = d.DocumentRepository.ExecuteInTransaction(func(tx *gorm.DB) error {
		if err := d.DocumentRepository.UpdateWithTransaction(tx, &document, documentUUID); err != nil {
			log.Printf("Error: %s", err)
			return fmt.Errorf("failed to update document details: %w", err)
		}

		if err := d.DocumentAccessRepository.UpdateWithTransaction(tx, &creatorAccess, creatorAccess.ID); err != nil {
			log.Printf("Error: %s", err)
			return fmt.Errorf("failed to update document details: %w", err)
		}

		if err := d.DocumentAccessRepository.UpdateWithTransaction(tx, &recipientAccess, recipientAccess.ID); err != nil {
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

func (d *DocumentController) InviteCollaborator(c *gin.Context) {
	var payload struct {
		DocumentId string     `json:"documentId"`
		Email      string     `json:"email"`
		Role       model.Role `json:"role"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

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

	documentUUID, err := uuid.Parse(payload.DocumentId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var document model.Document
	if err := d.DocumentRepository.GetOne(documentUUID, &document); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Document not found"})
			return
		}
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	token, err := utils.NewCodeGenerator().GenerateSecureToken(16)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating invite token"})
		return
	}

	var findUser model.User
	var collaboratorId *uuid.UUID
	if err := d.UserRepository.GetByEmail(&findUser, payload.Email); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Error retrieving user details: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	} else {
		collaboratorId = &findUser.ID
	}

	var oldInvite model.Invite
	if err := d.InviteRepository.GetOneByEmailAndDocId(&oldInvite, payload.Email, documentUUID); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	}

	if oldInvite.Status == model.InviteStatus(model.Accepted) {
		c.JSON(http.StatusOK, gin.H{"error": "Invite already accepted"})
		return
	}

	if oldInvite.Status == model.InviteStatus(model.Pending) {
		var deleteInvite model.Invite
		if err := d.InviteRepository.Delete(&deleteInvite, oldInvite.ID); err != nil {
			log.Printf("Error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	}

	newInvite := model.Invite{
		Email:          &payload.Email,
		DocumentId:     documentUUID,
		InviterId:      userDetails.ID,
		Role:           payload.Role,
		Status:         model.InviteStatus(model.Pending),
		Token:          token,
		CollaboratorId: collaboratorId,
	}

	if err := d.InviteRepository.Create(&newInvite); err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	envVars, err := constants.LoadEnv()
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	inviteUrl := fmt.Sprintf("%s/invite/%s", envVars.FE_ROOT_URL, token)
	err = handlers.SendMail(payload.Email, "invite", "Invite Mail", handlers.Invite{
		InviteLink:    inviteUrl,
		DocumentTitle: document.Title,
		Role:          payload.Role,
		FullName:      payload.Email,
		Year:          time.Now().Year(),
	})
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invite sent successfully"})
}

func (d *DocumentController) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")
	envVars, err := constants.LoadEnv()
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var invite model.Invite
	if err := d.InviteRepository.GetOneByToken(token, &invite); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Invitation not found"})
		}
	}

	var user model.User
	if err := d.UserRepository.GetByEmail(&user, *invite.Email); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error retrieving user details: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	} else {
		documentAccess := model.DocumentAccess{
			CollaboratorId: user.ID,
			DocumentId:     invite.DocumentId,
			Role:           invite.Role,
		}
		if err := d.DocumentAccessRepository.Create(&documentAccess); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error granting document access"})
			return
		}

		invite.Status = model.InviteStatus(model.Accepted)
		if err := d.InviteRepository.Update(&invite, invite.ID); err != nil {
			log.Printf("Error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		tokenGenerator, err := jwt.NewSession()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		accessToken, err := tokenGenerator.GenerateAccessToken(user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
			return
		}

		refreshToken, err := tokenGenerator.GenerateRefreshToken(user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Account setup complete",
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"redirectTo":   fmt.Sprintf("/get-document/%s", invite.DocumentId.String()),
		})

	}

	newUser := model.User{
		Email: *invite.Email,
	}
	createdUser, err := d.UserRepository.Create(&newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	var document model.Document
	if err := d.DocumentRepository.GetOne(invite.DocumentId, &document); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "document not found"})
			return
		}
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	invite.Status = model.InviteStatus(model.Accepted)
	if err := d.InviteRepository.Update(&invite, invite.ID); err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	accountSetupUrl := fmt.Sprintf("%s/complete-registration/%s?documentId=%s", envVars.DB_URI, createdUser.ID, invite.DocumentId)
	err = handlers.SendMail(user.Email, "welcome", "Welcome Mail", handlers.AccountSetup{
		DocumentTitle:    document.Title,
		Role:             invite.Role,
		AccountSetupLink: accountSetupUrl,
	})
	if err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "invite accepted"})
}

func (d *DocumentController) GenerateDocPDF(c *gin.Context) {
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

	documentId := c.Param("documentId")
	documentUUID, err := uuid.Parse(documentId)
	if err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var document model.Document
	if err := d.DocumentRepository.GetOne(documentUUID, &document); err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var documentMetaData model.DocumentMetadata
	if err := d.DocumentMetadataRepository.GetOneByDocId(documentUUID, &documentMetaData); err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	uploaded, err := utils.DocumentHandler(&document, &documentMetaData)
	if err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	documentMedia := model.DocumentMedia{
		DocumentID: document.ID,
		PublicID:   uploaded.PublicID,
		SecureURL:  uploaded.SecureURL,
		Format:     "pdf",
	}

	if err := d.DocumentMediaRepository.Create(&documentMedia); err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document generated", "documentLink": uploaded.SecureURL})
}
