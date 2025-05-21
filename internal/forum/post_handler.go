package forum

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/saqreed/golang-basic-forum/internal/auth"
	"github.com/saqreed/golang-basic-forum/internal/storage"
	"github.com/saqreed/golang-basic-forum/pkg/logger"
)

type PostResponse struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type PostHandler struct {
	postRepo *storage.PostRepository
	userRepo *storage.UserRepository
	auth     *auth.Service
}

func NewPostHandler(postRepo *storage.PostRepository, userRepo *storage.UserRepository, auth *auth.Service) *PostHandler {
	return &PostHandler{
		postRepo: postRepo,
		userRepo: userRepo,
		auth:     auth,
	}
}

func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetPosts(w, r)
	case http.MethodPost:
		h.handleCreatePost(w, r)
	case http.MethodDelete:
		h.handleDeletePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PostHandler) handleGetPosts(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			http.Error(w, "Invalid page parameter", http.StatusBadRequest)
			return
		}
	}

	pageSize := 10
	if pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			http.Error(w, "Invalid page_size parameter", http.StatusBadRequest)
			return
		}
	}

	// Get posts from database
	ctx := r.Context()
	posts, err := h.postRepo.GetPosts(ctx, pageSize, (page-1)*pageSize)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get posts")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert posts to response format
	response := make([]PostResponse, len(posts))
	for i, post := range posts {
		user, err := h.userRepo.GetUserByID(ctx, post.UserID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get user")
			continue
		}

		response[i] = PostResponse{
			ID:        post.ID,
			Title:     post.Title,
			Content:   post.Content,
			UserID:    post.UserID,
			Username:  user.Username,
			CreatedAt: post.CreatedAt.Format(time.RFC3339),
			UpdatedAt: post.UpdatedAt.Format(time.RFC3339),
		}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	resp, err := h.auth.ValidateToken(ctx, &proto.ValidateTokenRequest{Token: token})
	if err != nil || !resp.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create post
	post, err := h.postRepo.CreatePost(ctx, req.Title, req.Content, resp.UserId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create post")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get user info
	user, err := h.userRepo.GetUserByID(ctx, post.UserID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send response
	response := PostResponse{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		UserID:    post.UserID,
		Username:  user.Username,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
		UpdatedAt: post.UpdatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	resp, err := h.auth.ValidateToken(ctx, &proto.ValidateTokenRequest{Token: token})
	if err != nil || !resp.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get post ID from URL
	postIDStr := r.URL.Path[len("/api/posts/"):]
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Check if user is admin or post owner
	post, err := h.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	if post.UserID != resp.UserId && resp.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Delete post
	if err := h.postRepo.DeletePost(ctx, postID); err != nil {
		logger.Error().Err(err).Msg("Failed to delete post")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
