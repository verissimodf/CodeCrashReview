package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	"codecrashreview/backend/internal/config"
	_ "codecrashreview/backend/internal/database/migrations"
)

func main() {
	cfg := config.Load()

	// Parse flags
	dir := flag.String("dir", "./migrations", "migrations directory")
	command := flag.String("command", "up", "goose command (up, down, reset, status, version)")
	flag.Parse()

	// Connect to database
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Get standard database connection for goose
	db, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Failed to acquire connection: %v", err)
	}
	defer db.Release()

	// Set goose dialect
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}

	// Run goose command
	switch *command {
	case "up":
		if err := goose.Up(db.Conn(), *dir); err != nil {
			log.Fatalf("Failed to run migrations up: %v", err)
		}
		log.Println("Migrations applied successfully")
	case "down":
		if err := goose.Down(db.Conn(), *dir); err != nil {
			log.Fatalf("Failed to run migrations down: %v", err)
		}
		log.Println("Migration rolled back successfully")
	case "reset":
		if err := goose.Reset(db.Conn(), *dir); err != nil {
			log.Fatalf("Failed to reset migrations: %v", err)
		}
		log.Println("Migrations reset successfully")
	case "status":
		if err := goose.Status(db.Conn(), *dir); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
	case "version":
		if err := goose.Version(db.Conn(), *dir); err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}