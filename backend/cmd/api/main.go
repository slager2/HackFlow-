package main

import (
	"log/slog"
	"net/http"

	"hackflow-api/internal/config"
	"hackflow-api/internal/database"
	"hackflow-api/internal/handlers"
	"hackflow-api/internal/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Initialize Configuration
	cfg := config.Load()

	// 2. Initialize Structured Logger
	logger.Setup(cfg.Env)
	slog.Info("Starting HackFlow API Server", "env", cfg.Env, "port", cfg.Port)

	// 3. Initialize Database
	db, err := database.Init(cfg)
	if err != nil {
		slog.Error("Critical error: unable to initialize database", "error", err)
		return
	}

	// 4. Initialize HTTP Handlers with DB Dependency
	h := handlers.New(db)
	aiHandler := handlers.NewSearchAIHandler(cfg)

	// 5. Initialize Gin Router
	if cfg.Env == "production" || cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Configure CORS for Next.js frontend
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	corsConfig.AllowMethods = []string{"GET", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	r.Use(cors.New(corsConfig))

	// Defines API Routes
	api := r.Group("/api")
	{
		api.GET("/hackathons", h.GetHackathons)
		api.GET("/search", aiHandler.SearchAI)
	}

	// 6. Start the Server
	addr := ":" + cfg.Port
	slog.Info("Server listening", "address", addr)
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		slog.Error("Critical server error", "error", err)
	}
}
