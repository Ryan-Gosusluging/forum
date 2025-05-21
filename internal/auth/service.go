package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/saqreed/golang-basic-forum/internal/storage"
	"github.com/saqreed/golang-basic-forum/pkg/config"
	"github.com/saqreed/golang-basic-forum/pkg/logger"
	"github.com/saqreed/golang-basic-forum/pkg/proto"
)

type Service struct {
	proto.UnimplementedAuthServiceServer
	userRepo *storage.UserRepository
	cfg      *config.Config
}

func NewService(userRepo *storage.UserRepository, cfg *config.Config) *Service {
	return &Service{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *Service) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	// Check if user already exists
	_, err := s.userRepo.GetUserByUsername(ctx, req.Username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	// Create new user
	user, err := s.userRepo.CreateUser(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create user")
		return nil, errors.New("failed to create user")
	}

	return &proto.RegisterResponse{
		UserId:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

func (s *Service) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Verify password
	if err := s.userRepo.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate token")
		return nil, errors.New("failed to generate token")
	}

	return &proto.LoginResponse{
		Token:    tokenString,
		UserId:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}, nil
}

func (s *Service) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return &proto.ValidateTokenResponse{Valid: false}, nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &proto.ValidateTokenResponse{
			Valid:    true,
			UserId:   int64(claims["user_id"].(float64)),
			Username: claims["username"].(string),
		}, nil
	}

	return &proto.ValidateTokenResponse{Valid: false}, nil
}
