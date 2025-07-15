package ws

import (
	"encoding/json"
	"errors"
	"log"
	"realTimeEditor/internal/model"
	"realTimeEditor/internal/repositories"
	"realTimeEditor/pkg/jwt"

	"github.com/google/uuid"
	socketio "github.com/googollee/go-socket.io"
)

type SocketHandler struct {
	DocumentRepository       *repositories.DocumentRepository
	DocumentAccessRepository *repositories.DocumentAccessRepository
	SessionService           *jwt.Session
	UserRepository           *repositories.UserRepository
}

func NewSocketHandler(
	documentRepo *repositories.DocumentRepository,
	documentAccessRepo *repositories.DocumentAccessRepository,
	session *jwt.Session,
	userRepo *repositories.UserRepository,
) *SocketHandler {
	return &SocketHandler{
		DocumentRepository:       documentRepo,
		DocumentAccessRepository: documentAccessRepo,
		SessionService:           session,
		UserRepository:           userRepo,
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

		s.SetContext(map[string]string{
			"userId": user.ID.String(),
			"email":  user.Email,
		})

		log.Println("Connected:", s.ID())
		log.Println("Authenticated connection:", s.ID())
		s.Emit("connected", map[string]string{
			"message": "Connection established",
			"userId":  user.ID.String(),
		})
		return nil
	})

	server.OnEvent("/ws", "edit", func(s socketio.Conn, data map[string]interface{}) {
		ctx := s.Context().(map[string]string)
		userId := ctx["userId"]

		docId, ok := data["ID"].(string)
		if !ok || docId == "" {
			s.Emit("error", "Invalid document ID")
			return
		}

		hasAccess, err := sh.DocumentAccessRepository.HasEditAccess(userId, docId)
		if err != nil {
			s.Emit("error", "Error validating editor access")
			log.Printf("Access validation failed: %v", err)
			return
		}

		if !hasAccess {
			s.Emit("error", "You do not have access to edit this document")
			return
		}

		// content, ok := data["content"].(string)
		// if !ok {
		// 	s.Emit("error", "Invalid content")
		// 	return
		// }

		// log.Printf("User %s editing document %s", userId, docId)

		var document model.Document
		d, err := json.Marshal(data)
		if err != nil {
			log.Println("Failed to marshal message:", err)
			s.Emit("error", "Invalid message format")
			return
		}
		if err := json.Unmarshal(d, &document); err != nil {
			log.Println("Invalid post:", err)
			return
		}

		docUUID, err := uuid.Parse(docId)
		if err != nil {
			log.Println("Failed to parse doc uuid:", err)
			s.Emit("error", "Internal server error")
			return
		}

		//TODO: Address this later.
		// type EditPayload struct {
		// 	DocumentID string          `json:"documentId"`
		// 	Content    string          `json:"content"`          // updated content or diff/patch
		// 	Version    int             `json:"version"`          // optional
		// 	Cursor     *CursorPosition `json:"cursor,omitempty"` // optional
		// }

		if err := sh.DocumentRepository.Update(&document, docUUID); err != nil {
			log.Println("Failed to update document:", err)
			s.Emit("error", "Failed to update document")
			return
		}
	})

}
