package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Ryan-Gosusluging/forum/internal/auth"
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
	userRepo := storage.NewUserRepository(db)

	// Create auth service
	authService := auth.NewService(userRepo, cfg)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	proto.RegisterAuthServiceServer(grpcServer, authService)

	// Start listening
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.AuthServicePort))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to listen")
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info().Msg("Shutting down server...")
		grpcServer.GracefulStop()
	}()

	logger.Info().Msgf("Auth service starting on port %d", cfg.AuthServicePort)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal().Err(err).Msg("Failed to serve")
	}
}
