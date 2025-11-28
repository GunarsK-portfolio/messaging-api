package repository

import (
	"context"
	"fmt"

	"github.com/GunarsK-portfolio/portfolio-common/models"
	"gorm.io/gorm"
)

// CreateContactMessage creates a new contact message
func (r *repository) CreateContactMessage(ctx context.Context, message *models.ContactMessage) error {
	err := r.db.WithContext(ctx).
		Omit("ID", "CreatedAt", "UpdatedAt").
		Create(message).Error
	if err != nil {
		return fmt.Errorf("failed to create contact message: %w", err)
	}
	return nil
}

// GetContactMessages retrieves all contact messages
func (r *repository) GetContactMessages(ctx context.Context) ([]models.ContactMessage, error) {
	var messages []models.ContactMessage
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get contact messages: %w", err)
	}
	return messages, nil
}

// GetContactMessageByID retrieves a contact message by ID
func (r *repository) GetContactMessageByID(ctx context.Context, id int64) (*models.ContactMessage, error) {
	var message models.ContactMessage
	err := r.db.WithContext(ctx).First(&message, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get contact message by id %d: %w", id, err)
	}
	return &message, nil
}

// UpdateContactMessageStatus updates the status of a contact message
func (r *repository) UpdateContactMessageStatus(ctx context.Context, id int64, status string, lastError *string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if lastError != nil {
		updates["last_error"] = *lastError
	}
	if status == models.MessageStatusSent {
		updates["sent_at"] = r.db.NowFunc()
	}
	if status == models.MessageStatusFailed {
		updates["attempts"] = gorm.Expr("attempts + 1")
	}

	result := r.db.WithContext(ctx).
		Model(&models.ContactMessage{}).
		Where("id = ?", id).
		Updates(updates)

	if err := checkRowsAffected(result); err != nil {
		return fmt.Errorf("failed to update contact message status: %w", err)
	}
	return nil
}
