package handlers

import (
	"github.com/GunarsK-templates/template-api/internal/repository"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	repo repository.Repository
}

// New creates a new Handler instance
func New(repo repository.Repository) *Handler {
	return &Handler{repo: repo}
}
