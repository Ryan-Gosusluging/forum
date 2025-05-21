package storage

import (
	"context"
	"database/sql"
	"time"
)

// Comment represents a comment in the forum
type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CommentRepository defines the interface for comment operations
type CommentRepository interface {
	GetCommentsByPostID(ctx context.Context, postID int64) ([]Comment, error)
	GetCommentByID(ctx context.Context, id int64) (*Comment, error)
	CreateComment(ctx context.Context, comment *Comment) error
	DeleteComment(ctx context.Context, id int64) error
}

// CommentRepositoryImpl implements CommentRepository
type CommentRepositoryImpl struct {
	db *DB
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *DB) CommentRepository {
	return &CommentRepositoryImpl{db: db}
}

// GetCommentsByPostID retrieves all comments for a specific post
func (r *CommentRepositoryImpl) GetCommentsByPostID(ctx context.Context, postID int64) ([]Comment, error) {
	query := `
		SELECT id, post_id, content, user_id, created_at, updated_at
		FROM comments
		WHERE post_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.Content,
			&comment.UserID,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// GetCommentByID retrieves a comment by its ID
func (r *CommentRepositoryImpl) GetCommentByID(ctx context.Context, id int64) (*Comment, error) {
	query := `
		SELECT id, post_id, content, user_id, created_at, updated_at
		FROM comments
		WHERE id = $1
	`

	var comment Comment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.Content,
		&comment.UserID,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

// CreateComment creates a new comment
func (r *CommentRepositoryImpl) CreateComment(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments (post_id, content, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		comment.PostID,
		comment.Content,
		comment.UserID,
		comment.CreatedAt,
		comment.UpdatedAt,
	).Scan(&comment.ID)

	return err
}

// DeleteComment deletes a comment by its ID
func (r *CommentRepositoryImpl) DeleteComment(ctx context.Context, id int64) error {
	query := `
		DELETE FROM comments
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
