package router

import (
	"realTimeEditor/internal/controllers"
	"realTimeEditor/internal/middlewares"
	"realTimeEditor/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func DocumentMetadataRouter(g *gin.Engine, d *controllers.DocumentMetadataController, m *middlewares.AuthMiddleware, s *jwt.Session) {
	documentMetadataGroup := g.Group("/document-metadata")
	documentMetadataGroup.Use(m.UserAuth(s))
	{
		documentMetadataGroup.POST("/create", d.Create)
		documentMetadataGroup.GET("/get-one", d.GetDocumentMetadata)
		documentMetadataGroup.PATCH("/update", d.Update)
		documentMetadataGroup.DELETE("/delete", d.Delete)
	}
}
