package router

import (
	"realTimeEditor/internal/controllers"
	"realTimeEditor/internal/middlewares"
	"realTimeEditor/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func UserRouter(g *gin.Engine, u *controllers.UserController, m *middlewares.AuthMiddleware, s *jwt.Session) {
	authGroup := g.Group("/auth")
	{
		authGroup.POST("/register", u.Create)
		authGroup.POST("/login", u.Login)
		authGroup.POST("/forgot-password", u.ForgotPassword)
		authGroup.POST("/verify-reset-code", u.VerifyResetCode)
		authGroup.POST("/access-token", u.GenerateAccessToken)
		authGroup.POST("/complete-account", u.CompleteAccount)
	}

	authGroup.Use(m.UserAuth(s))
	{
		authGroup.POST("/reset-password", u.ResetPassword)
	}

	userGroup := g.Group("/member")
	userGroup.Use(m.UserAuth(s))
	{
		userGroup.GET("/profile", u.Profile)
		userGroup.POST("/profile-upload", u.UploadProfilePicture)
	}
}
