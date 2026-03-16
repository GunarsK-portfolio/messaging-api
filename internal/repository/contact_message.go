package repository

import (
	"context"
	"fmt"

	"github.com/GunarsK-portfolio/portfolio-common/models"
	"gorm.io/gorm"
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

// GetEmails retrieves all emails
func (r *repository) GetEmails(ctx context.Context) ([]models.Email, error) {
	var emails []models.Email
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
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

// UpdateEmailStatus updates the status of an email
func (r *repository) UpdateEmailStatus(ctx context.Context, id int64, status string, lastError *string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if lastError != nil {
		updates["last_error"] = *lastError
	}
	if status == models.EmailStatusSent {
		updates["sent_at"] = r.db.NowFunc()
	}
	if status == models.EmailStatusFailed {
		updates["attempts"] = gorm.Expr("attempts + 1")
	}

	result := r.db.WithContext(ctx).
		Model(&models.Email{}).
		Where("id = ?", id).
		Updates(updates)

	if err := checkRowsAffected(result); err != nil {
		return fmt.Errorf("failed to update email status: %w", err)
	}
	return nil
}
