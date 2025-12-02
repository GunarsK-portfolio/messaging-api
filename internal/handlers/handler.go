package handlers

import (
	"github.com/GunarsK-portfolio/messaging-api/internal/repository"
	commonhandlers "github.com/GunarsK-portfolio/portfolio-common/handlers"
	"github.com/GunarsK-portfolio/portfolio-common/queue"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	repo      repository.Repository
	publisher queue.Publisher
}

// New creates a new Handler instance
func New(repo repository.Repository, publisher queue.Publisher) *Handler {
	return &Handler{
		repo:      repo,
		publisher: publisher,
	}
}

// setLocationHeader wraps the common helper
var setLocationHeader = commonhandlers.SetLocationHeader
