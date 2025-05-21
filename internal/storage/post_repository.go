package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Post struct {
	ID        int64
	Title     string
	Content   string
	UserID    int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PostRepository struct {
	db *DB
}

func NewPostRepository(db *DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) CreatePost(ctx context.Context, title, content string, userID int64) (*Post, error) {
	query := `
		INSERT INTO posts (title, content, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, title, content, user_id, created_at, updated_at
	`

	post := &Post{}
	err := r.db.QueryRowContext(ctx, query, title, content, userID).
		Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return post, nil
}

func (r *PostRepository) GetPostByID(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT id, title, content, user_id, created_at, updated_at
		FROM posts
		WHERE id = $1
	`

	post := &Post{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	return post, nil
}

func (r *PostRepository) GetPosts(ctx context.Context, limit, offset int) ([]*Post, error) {
	query := `
		SELECT id, title, content, user_id, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *PostRepository) DeletePost(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("post not found")
	}

	return nil
}
