package ws

import (
	"errors"
	"realTimeEditor/internal/repositories"

	socketio "github.com/googollee/go-socket.io"
)

type SocketHandler struct {
	DocumentRepository       *repositories.DocumentRepository
	DocumentAccessRepository *repositories.DocumentAccessRepository
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

		return nil
	})
}
