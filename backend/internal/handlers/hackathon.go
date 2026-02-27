package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"hackflow-api/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler contains injected dependencies for the HTTP handlers
type Handler struct {
	DB *gorm.DB
}

// New creates a new Handler with the given database connection
func New(db *gorm.DB) *Handler {
	return &Handler{
		DB: db,
	}
}

// GetHackathons handles the GET /api/hackathons requests
func (h *Handler) GetHackathons(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))

	var hackathons []models.Hackathon

	if query == "" {
		if err := h.DB.Find(&hackathons).Error; err != nil {
			slog.Error("Failed to fetch hackathons from database", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
			return
		}
	} else {
		searchPattern := "%" + query + "%"
		slog.Debug("Searching hackathons", "query", query)
		if err := h.DB.Where("title ILIKE ? OR city ILIKE ?", searchPattern, searchPattern).Find(&hackathons).Error; err != nil {
			slog.Error("Failed to search hackathons", "error", err, "query", query)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search data"})
			return
		}
	}

	if hackathons == nil {
		hackathons = make([]models.Hackathon, 0)
	}

	c.JSON(http.StatusOK, hackathons)
}
