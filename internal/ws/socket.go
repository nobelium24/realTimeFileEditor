package ws

import (
	"errors"
	"log"
	"realTimeEditor/internal/model"
	"realTimeEditor/internal/repositories"
	"realTimeEditor/pkg/jwt"

	socketio "github.com/googollee/go-socket.io"
)

type SocketHandler struct {
	DocumentRepository       *repositories.DocumentRepository
	DocumentAccessRepository *repositories.DocumentAccessRepository
	SessionService           *jwt.Session
	UserRepository           *repositories.UserRepository
}

func NewSocketHandler(documentRepo *repositories.DocumentRepository,
	documentAccessRepo *repositories.DocumentAccessRepository) *SocketHandler {
	return &SocketHandler{
		DocumentRepository:       documentRepo,
		DocumentAccessRepository: documentAccessRepo,
	}
}

func (sh *SocketHandler) RegisterEvents(server *socketio.Server) {
	server.OnConnect("/ws", func(s socketio.Conn) error {
		if s == nil {
			return errors.New("nil connection")
		}
		token := s.RemoteHeader().Get("Authorization")
		if token == "" {
			return errors.New("authentication required")
		}

		email, err := sh.SessionService.VerifyAccessToken(token)
		if err != nil {
			s.Emit("error", "Invalid or expired session")
			log.Printf("Token validation failed: %v", err)
			return errors.New("authentication failed")
		}

		var user model.User
		if err := sh.UserRepository.GetByEmail(&user, email); err != nil {
			s.Emit("error", "Invalid or expired session")
			log.Printf("Token validation failed: %v", err)
			return errors.New("authentication failed")
		}

		log.Println("Connected:", s.ID())
		log.Println("Authenticated connection:", s.ID())
		s.Emit("connected", map[string]string{
			"message": "Connection established",
			"userId":  user.ID.String(),
		})
		return nil
	})
}
