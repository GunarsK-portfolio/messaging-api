package repository

import (
	"context"
	"fmt"

	"github.com/GunarsK-portfolio/portfolio-common/models"
	commonrepo "github.com/GunarsK-portfolio/portfolio-common/repository"
)

// CreateEmail creates a new email record
func (r *repository) CreateEmail(ctx context.Context, email *models.Email) error {
	err := r.db.WithContext(ctx).
		Omit("ID", "CreatedAt", "UpdatedAt").
		Create(email).Error
	if err != nil {
		return fmt.Errorf("failed to create email: %w", err)
	}
	return nil
}

// defaultEmailLimit caps the number of emails returned to prevent OOM on large datasets
const defaultEmailLimit = 100

// GetEmails retrieves recent emails (capped at defaultEmailLimit)
func (r *repository) GetEmails(ctx context.Context) ([]models.Email, error) {
	var emails []models.Email
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(defaultEmailLimit).
		Find(&emails).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get emails: %w", err)
	}
	return emails, nil
}

// GetEmailByID retrieves an email by ID
func (r *repository) GetEmailByID(ctx context.Context, id int64) (*models.Email, error) {
	var email models.Email
	err := r.db.WithContext(ctx).First(&email, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get email by id %d: %w", id, err)
	}
	return &email, nil
}

// UpdateEmailStatus delegates to the shared helper in portfolio-common
func (r *repository) UpdateEmailStatus(ctx context.Context, id int64, status string, lastError *string) error {
	return commonrepo.UpdateEmailStatus(r.db, ctx, id, status, lastError)
}
