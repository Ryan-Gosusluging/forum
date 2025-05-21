package forum

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Ryan-Gosusluging/forum/internal/auth"
	"github.com/Ryan-Gosusluging/forum/internal/storage"
	"github.com/Ryan-Gosusluging/forum/pkg/logger"
)

// CommentHandler handles HTTP requests related to forum comments
type CommentHandler struct {
	commentRepo storage.CommentRepository
	authService *auth.Service
	logger      *logger.Logger
}

// NewCommentHandler creates a new CommentHandler instance
func NewCommentHandler(commentRepo storage.CommentRepository, authService *auth.Service, logger *logger.Logger) *CommentHandler {
	return &CommentHandler{
		commentRepo: commentRepo,
		authService: authService,
		logger:      logger,
	}
}

// handleGetComments retrieves comments for a specific post
func (h *CommentHandler) handleGetComments(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseInt(r.URL.Query().Get("post_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	comments, err := h.commentRepo.GetCommentsByPostID(r.Context(), postID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get comments")
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// handleCreateComment creates a new comment
func (h *CommentHandler) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authService.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var comment storage.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	comment.UserID = userID
	comment.CreatedAt = time.Now()

	if err := h.commentRepo.CreateComment(r.Context(), &comment); err != nil {
		h.logger.Error().Err(err).Msg("Failed to create comment")
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// handleDeleteComment deletes a comment
func (h *CommentHandler) handleDeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authService.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	commentID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	comment, err := h.commentRepo.GetCommentByID(r.Context(), commentID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get comment")
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if comment.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.commentRepo.DeleteComment(r.Context(), commentID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to delete comment")
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
