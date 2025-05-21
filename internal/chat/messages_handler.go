package chat

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/yourusername/golang-basic-forum/pkg/logger"
)

type MessageResponse struct {
	ID        int64   `json:"id"`
	UserID    *int64  `json:"user_id,omitempty"`
	Username  *string `json:"username,omitempty"`
	Content   string  `json:"content"`
	CreatedAt string  `json:"created_at"`
}

type MessagesHandler struct {
	hub *Hub
}

func NewMessagesHandler(hub *Hub) *MessagesHandler {
	return &MessagesHandler{
		hub: hub,
	}
}

func (h *MessagesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры запроса
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // значение по умолчанию
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	// Получаем сообщения из базы данных
	ctx := r.Context()
	messages, err := h.hub.chatRepo.GetRecentMessages(ctx, limit)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get messages")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Преобразуем сообщения в формат ответа
	response := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		response[i] = MessageResponse{
			ID:        msg.ID,
			UserID:    msg.UserID,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
		}
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
