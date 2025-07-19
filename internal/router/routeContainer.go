package router

import (
	"realTimeEditor/internal/controllers"
	"realTimeEditor/internal/middlewares"
	"realTimeEditor/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type RouterContainer struct {
	UserController     *controllers.UserController
	DocumentController *controllers.DocumentController
	AuthMiddleware     *middlewares.AuthMiddleware
	Session            *jwt.Session
}

func (rc *RouterContainer) Register(r *gin.Engine) {
	UserRouter(r, rc.UserController, rc.AuthMiddleware, rc.Session)
	DocumentRouter(r, rc.DocumentController, rc.AuthMiddleware, rc.Session)
}
