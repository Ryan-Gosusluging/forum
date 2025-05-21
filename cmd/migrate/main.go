package main

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/Ryan-Gosusluging/forum/pkg/config"
)

func main() {
	cfg := config.NewConfig()
	dbURL := cfg.GetDBURL()

	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalf("Error creating migrate instance: %v", err)
	}

	if len(os.Args) < 2 {
		log.Fatal("Please specify command: up, down, or version")
	}

	cmd := os.Args[1]
	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Error applying migrations: %v", err)
		}
		log.Println("Migrations applied successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Error rolling back migrations: %v", err)
		}
		log.Println("Migrations rolled back successfully")
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Error getting version: %v", err)
		}
		log.Printf("Current version: %d, dirty: %v", version, dirty)
	default:
		log.Fatal("Unknown command. Please use: up, down, or version")
	}
}
