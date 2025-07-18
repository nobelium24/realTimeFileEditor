package router

import (
	"realTimeEditor/internal/controllers"
	"realTimeEditor/internal/middlewares"
	"realTimeEditor/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func DocumentRouter(g *gin.Engine, d *controllers.DocumentController, m *middlewares.AuthMiddleware, s *jwt.Session) {
	documentGroup := g.Group("/document")
	documentGroup.Use(m.UserAuth(s))
	{
		documentGroup.POST("/create", d.Create)
		documentGroup.GET("/user-created-docs", d.GetUserCreatedDocuments)
		documentGroup.GET("/get-one", d.GetSingleDocument)
		documentGroup.DELETE("/revoke-access/:documentAccessId", d.RevokeAccess)
		documentGroup.DELETE("/delete/:id", d.DeleteDocument)
		documentGroup.PATCH("/modify-access/:documentAccessId/:newRole", d.ModifyAccess)
		documentGroup.GET("/all", d.FetchAllDocuments)
		documentGroup.GET("/collaborators/:id", d.FetchCollaborators)
		documentGroup.PATCH("/transfer-ownership/:documentId/:recipientId", d.TransferOwnership)
		documentGroup.POST("/invite-collaborator", d.InviteCollaborator)
	}

	docGroup := g.Group("/invite")
	{
		docGroup.GET("/accept", d.AcceptInvitation)
	}
}
