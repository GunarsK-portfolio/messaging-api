package handlers

import (
	"github.com/GunarsK-portfolio/messaging-api/internal/repository"
	commonhandlers "github.com/GunarsK-portfolio/portfolio-common/handlers"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	repo repository.Repository
}

// New creates a new Handler instance
func New(repo repository.Repository) *Handler {
	return &Handler{repo: repo}
}

// setLocationHeader wraps the common helper for backward compatibility
var setLocationHeader = commonhandlers.SetLocationHeader
