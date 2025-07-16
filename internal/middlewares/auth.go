package middlewares

import (
	"net/http"
	"realTimeEditor/internal/model"
	"realTimeEditor/internal/repositories"
	"realTimeEditor/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	UserRepository *repositories.UserRepository
}

func (a *AuthMiddleware) UserAuth(sessionService *jwt.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		email, err := sessionService.VerifyAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		var user model.User

		err = a.UserRepository.GetByEmail(&user, email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
