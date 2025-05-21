package storage

import (
	"context"
	"time"
)

type ChatMessage struct {
	ID        int64
	UserID    *int64
	Content   string
	CreatedAt time.Time
}

type ChatRepository struct {
	db *DB
}

func NewChatRepository(db *DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateMessage(ctx context.Context, userID *int64, content string) (*ChatMessage, error) {
	query := `
		INSERT INTO chat_messages (user_id, content)
		VALUES ($1, $2)
		RETURNING id, user_id, content, created_at
	`

	message := &ChatMessage{}
	err := r.db.QueryRowContext(ctx, query, userID, content).
		Scan(&message.ID, &message.UserID, &message.Content, &message.CreatedAt)

	if err != nil {
		return nil, err
	}

	return message, nil
}

func (r *ChatRepository) GetRecentMessages(ctx context.Context, limit int) ([]*ChatMessage, error) {
	query := `
		SELECT id, user_id, content, created_at
		FROM chat_messages
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*ChatMessage
	for rows.Next() {
		message := &ChatMessage{}
		err := rows.Scan(&message.ID, &message.UserID, &message.Content, &message.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *ChatRepository) DeleteOldMessages(ctx context.Context, olderThan time.Time) error {
	query := `DELETE FROM chat_messages WHERE created_at < $1`
	_, err := r.db.ExecContext(ctx, query, olderThan)
	return err
}
