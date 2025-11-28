package repository

import (
	"context"
	"fmt"

	"github.com/GunarsK-portfolio/portfolio-common/models"
)

// GetAllRecipients retrieves all recipients
func (r *repository) GetAllRecipients(ctx context.Context) ([]models.Recipient, error) {
	var recipients []models.Recipient
	err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&recipients).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all recipients: %w", err)
	}
	return recipients, nil
}

// GetActiveRecipients retrieves only active recipients
func (r *repository) GetActiveRecipients(ctx context.Context) ([]models.Recipient, error) {
	var recipients []models.Recipient
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&recipients).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active recipients: %w", err)
	}
	return recipients, nil
}

// GetRecipientByID retrieves a recipient by ID
func (r *repository) GetRecipientByID(ctx context.Context, id int64) (*models.Recipient, error) {
	var recipient models.Recipient
	err := r.db.WithContext(ctx).First(&recipient, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get recipient by id %d: %w", id, err)
	}
	return &recipient, nil
}

// CreateRecipient creates a new recipient
func (r *repository) CreateRecipient(ctx context.Context, recipient *models.Recipient) error {
	err := r.db.WithContext(ctx).
		Omit("ID", "CreatedAt", "UpdatedAt").
		Create(recipient).Error
	if err != nil {
		return fmt.Errorf("failed to create recipient: %w", err)
	}
	return nil
}

// UpdateRecipient updates an existing recipient
func (r *repository) UpdateRecipient(ctx context.Context, recipient *models.Recipient) error {
	if err := r.safeUpdate(ctx, recipient, recipient.ID); err != nil {
		return fmt.Errorf("failed to update recipient: %w", err)
	}
	return nil
}

// DeleteRecipient deletes a recipient by ID
func (r *repository) DeleteRecipient(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Recipient{}, id)
	if err := checkRowsAffected(result); err != nil {
		return fmt.Errorf("failed to delete recipient: %w", err)
	}
	return nil
}
