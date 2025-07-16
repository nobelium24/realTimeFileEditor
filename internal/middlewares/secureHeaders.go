package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

func SecureHeadersMiddleware() gin.HandlerFunc {
	secureMiddleware := secure.New(secure.Options{
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
		STSSeconds:           31536000,
		STSIncludeSubdomains: true,
		SSLRedirect:          false, // set true in production
	})
	return func(c *gin.Context) {
		err := secureMiddleware.Process(c.Writer, c.Request)
		if err != nil {
			c.Abort()
			return
		}
		c.Next()
	}
}
