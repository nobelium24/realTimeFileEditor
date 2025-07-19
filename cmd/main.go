package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"realTimeEditor/config"
	"realTimeEditor/internal/controllers"
	"realTimeEditor/internal/jobs"
	"realTimeEditor/internal/middlewares"
	"realTimeEditor/internal/repositories"
	"realTimeEditor/internal/router"
	"realTimeEditor/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func CreateRouter(container *router.RouterContainer) *gin.Engine {
	r := gin.Default()
	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.RateLimiterMiddleware())
	r.Use(middlewares.SecureHeadersMiddleware())
	r.Use(middlewares.CSPMiddleware())
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 100<<20) // 1 MB
		c.Next()
	})

	container.Register(r)
	return r
}

func main() {
	// envVars, err := constants.LoadEnv()
	// if err != nil {
	// 	log.Printf("Error loading environment variables: %v\n", err)
	// 	return
	// }

	config.ConnectToDB()
	if config.DB == nil {
		log.Fatal("Database connection is not initialized")
	}
	log.Println("Database connection initialized successfully")

	defer config.CloseDB()

	userRepository := repositories.NewUserRepository(config.DB)
	documentRepository := repositories.NewDocumentRepository(config.DB)
	documentAccessRepository := repositories.NewDocumentAccessRepository(config.DB)
	forgotPasswordRepository := repositories.NewForgotPasswordRepository(config.DB)
	inviteRepository := repositories.NewInviteRepository(config.DB)
	documentMetaDataRepository := repositories.NewDocumentMetaDataRepository(config.DB)
	documentMediaRepository := repositories.NewDocumentMediaRepository(config.DB)

	userController := controllers.NewUserHandler(userRepository, forgotPasswordRepository)
	documentController := controllers.NewDocumentController(
		documentRepository, documentAccessRepository, inviteRepository, userRepository, documentMetaDataRepository,
		documentMediaRepository,
	)
	documentMetaDataController := controllers.NewDocumentMetaDataController(documentRepository, documentMetaDataRepository)

	// Middleware
	authMiddleware := &middlewares.AuthMiddleware{
		UserRepository: userRepository,
	}

	// Session
	session, err := jwt.NewSession()
	if err != nil {
		log.Fatalf("Error initializing session: %s", err)
	}

	container := router.RouterContainer{
		UserController:             userController,
		DocumentController:         documentController,
		DocumentMetadataController: documentMetaDataController,
		AuthMiddleware:             authMiddleware,
		Session:                    session,
	}

	router := CreateRouter(&container)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleanUpJob := jobs.NewReceiptCleanup(*documentMediaRepository)
	go cleanUpJob.Start(ctx)

	server := &http.Server{
		Addr:    ":9091",
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Println("Starting server on port 9091...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
