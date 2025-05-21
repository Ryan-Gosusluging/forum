package chat

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/saqreed/golang-basic-forum/internal/auth"
	"github.com/saqreed/golang-basic-forum/pkg/logger"
	"github.com/saqreed/golang-basic-forum/pkg/proto"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // В продакшене нужно настроить правильную проверку origin
	},
}

type Handler struct {
	hub  *Hub
	auth *auth.Service
}

func NewHandler(hub *Hub, auth *auth.Service) *Handler {
	return &Handler{
		hub:  hub,
		auth: auth,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Проверяем токен аутентификации
	token := r.URL.Query().Get("token")
	var userID *int64
	var username *string

	if token != "" {
		ctx := context.Background()
		resp, err := h.auth.ValidateToken(ctx, &proto.ValidateTokenRequest{Token: token})
		if err == nil && resp.Valid {
			userID = &resp.UserId
			username = &resp.Username
		}
	}

	// Обновляем соединение до WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to upgrade connection")
		return
	}

	// Создаем нового клиента
	client := &Client{
		hub:      h.hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		username: username,
	}

	// Регистрируем клиента
	h.hub.register <- client

	// Запускаем горутины для чтения и записи
	go client.writePump()
	go client.readPump()
}
