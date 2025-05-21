package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Ryan-Gosusluging/forum/internal/auth"
	"github.com/Ryan-Gosusluging/forum/internal/chat"
	"github.com/Ryan-Gosusluging/forum/internal/storage"
	"github.com/Ryan-Gosusluging/forum/pkg/config"
	"github.com/Ryan-Gosusluging/forum/pkg/logger"
	"google.golang.org/grpc"
)

func main() {
	// Initialize logger
	logger.Init()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create database connection
	db, err := storage.NewDB(context.Background(), cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Create repositories
	chatRepo := storage.NewChatRepository(db)

	// Create auth service connection
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(cfg.AuthServicePort), grpc.WithInsecure())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to auth service")
	}
	defer conn.Close()

	authClient := proto.NewAuthServiceClient(conn)
	authService := auth.NewService(nil, cfg) // We don't need user repo here

	// Create chat hub
	chatHub := chat.NewHub(chatRepo, cfg)
	go chatHub.Run(context.Background())

	// Create HTTP handlers
	chatHandler := chat.NewHandler(chatHub, authService)
	messagesHandler := chat.NewMessagesHandler(chatHub)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle("/ws", chatHandler)
	mux.Handle("/api/chat/messages", messagesHandler)

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.ForumServicePort),
		Handler: mux,
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info().Msg("Shutting down server...")
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Error().Err(err).Msg("Error during server shutdown")
		}
	}()

	// Start HTTP server
	logger.Info().Msgf("Forum service starting on port %d", cfg.ForumServicePort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("Failed to start HTTP server")
	}
}
