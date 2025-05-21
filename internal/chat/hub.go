package chat

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/Ryan-Gosusluging/forum/internal/storage"
	"github.com/Ryan-Gosusluging/forum/pkg/config"
	"github.com/Ryan-Gosusluging/forum/pkg/logger"
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   *int64
	username *string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	chatRepo   *storage.ChatRepository
	cfg        *config.Config
}

func NewHub(chatRepo *storage.ChatRepository, cfg *config.Config) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		chatRepo:   chatRepo,
		cfg:        cfg,
	}
}

func (h *Hub) Run(ctx context.Context) {
	// Start cleanup goroutine
	go h.cleanupOldMessages(ctx)

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) cleanupOldMessages(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			olderThan := time.Now().Add(-h.cfg.ChatMessageTTL)
			if err := h.chatRepo.DeleteOldMessages(ctx, olderThan); err != nil {
				logger.Error().Err(err).Msg("Failed to cleanup old messages")
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error().Err(err).Msg("WebSocket read error")
			}
			break
		}

		// Store message in database
		ctx := context.Background()
		_, err = c.hub.chatRepo.CreateMessage(ctx, c.userID, string(message))
		if err != nil {
			logger.Error().Err(err).Msg("Failed to store message")
			continue
		}

		// Broadcast message to all clients
		c.hub.broadcast <- []byte(message)
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
