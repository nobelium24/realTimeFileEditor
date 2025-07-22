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
	"realTimeEditor/internal/ws"
	"realTimeEditor/pkg/jwt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/socket.io/") {
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
		}
		h.ServeHTTP(w, r)
	})
}

func CreateRouter(container *router.RouterContainer) *gin.Engine {
	r := gin.Default()

	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.RateLimiterMiddleware())
	r.Use(middlewares.SecureHeadersMiddleware())
	r.Use(middlewares.CSPMiddleware())
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 100<<20) // Limit body size to 100MB
		c.Next()
	})

	// Serve Swagger
	r.Static("/swagger-ui", "./api/swagger-ui/dist")
	r.StaticFile("/swagger.yaml", "./api/swagger.yaml")
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger-ui/index.html")
	})

	container.Register(r)
	return r
}

func main() {
	// Step 1: Initialize DB
	config.ConnectToDB()
	if config.DB == nil {
		log.Fatal("Database connection is not initialized")
	}
	defer config.CloseDB()
	log.Println("Database connected")

	// Step 2: Set up repositories
	userRepo := repositories.NewUserRepository(config.DB)
	docRepo := repositories.NewDocumentRepository(config.DB)
	docAccessRepo := repositories.NewDocumentAccessRepository(config.DB)
	forgotPwdRepo := repositories.NewForgotPasswordRepository(config.DB)
	inviteRepo := repositories.NewInviteRepository(config.DB)
	docMetaRepo := repositories.NewDocumentMetaDataRepository(config.DB)
	docMediaRepo := repositories.NewDocumentMediaRepository(config.DB)

	// Step 3: Initialize controllers
	userCtrl := controllers.NewUserHandler(userRepo, forgotPwdRepo)
	docCtrl := controllers.NewDocumentController(docRepo, docAccessRepo, inviteRepo, userRepo, docMetaRepo, docMediaRepo)
	docMetaCtrl := controllers.NewDocumentMetaDataController(docRepo, docMetaRepo)

	// Step 4: Auth middleware & session service
	authMiddleware := &middlewares.AuthMiddleware{UserRepository: userRepo}
	sessionService, err := jwt.NewSession()
	if err != nil {
		log.Fatalf("Error initializing session: %s", err)
	}

	// Step 5: Set up router
	container := router.RouterContainer{
		UserController:             userCtrl,
		DocumentController:         docCtrl,
		DocumentMetadataController: docMetaCtrl,
		AuthMiddleware:             authMiddleware,
		Session:                    sessionService,
	}
	apiRouter := CreateRouter(&container)

	// Step 6: WebSocket server setup
	socketServer := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&websocket.Transport{
				CheckOrigin: func(r *http.Request) bool {
					return true // allow all for dev
				},
			},
		},
		PingTimeout:  60 * time.Second,
		PingInterval: 25 * time.Second,
	})

	// Step 7: Register socket events
	socketHandler := ws.NewSocketHandler(docRepo, docAccessRepo, sessionService, userRepo)
	socketHandler.RegisterEvents(socketServer)

	// Error handler for socket server
	socketServer.OnError("/", func(conn socketio.Conn, err error) {
		if conn != nil {
			log.Printf("Socket error [%s]: %v", conn.ID(), err)
			conn.Emit("error", err.Error())
		} else {
			log.Printf("Socket error [nil conn]: %v", err)
		}
	})

	// Serve socket server
	go func() {
		if err := socketServer.Serve(); err != nil {
			log.Fatalf("SocketIO server error: %v", err)
		}
	}()
	defer socketServer.Close()

	// Step 8: Setup background cleanup job
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cleanUpJob := jobs.NewReceiptCleanup(*docMediaRepo)
	go cleanUpJob.Start(ctx)

	// Step 9: Compose final HTTP server with both API and WS
	mux := http.NewServeMux()
	mux.Handle("/socket.io/", allowCORS(socketServer))
	mux.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	mux.Handle("/", apiRouter)

	server := &http.Server{
		Addr:    ":9091",
		Handler: mux,
	}

	// Step 10: Start server in goroutine
	go func() {
		log.Println("Starting server on port 9091...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Step 11: Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
