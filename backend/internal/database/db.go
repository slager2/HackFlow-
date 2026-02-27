package database

import (
	"fmt"
	"log/slog"

	"hackflow-api/internal/config"
	"hackflow-api/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Init connects to the PostgreSQL database and runs auto-migrations.
func Init(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err, "host", cfg.DBHost)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("Successfully connected to PostgreSQL", "host", cfg.DBHost, "db", cfg.DBName)

	slog.Info("Running auto-migrations")
	err = db.AutoMigrate(&models.Hackathon{})
	if err != nil {
		slog.Error("Failed to run migrations", "error", err)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Database schema synchronized")
	return db, nil
}
