package repository

import (
	"context"
	"fmt"

	"github.com/GunarsK-templates/template-api/internal/models"
)

// GetAllItems retrieves all items
func (r *repository) GetAllItems(ctx context.Context) ([]models.Item, error) {
	var items []models.Item
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all items: %w", err)
	}
	return items, nil
}

// GetItemByID retrieves an item by its ID
func (r *repository) GetItemByID(ctx context.Context, id int64) (*models.Item, error) {
	var item models.Item
	err := r.db.WithContext(ctx).
		First(&item, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get item by id %d: %w", id, err)
	}
	return &item, nil
}

// CreateItem creates a new item
func (r *repository) CreateItem(ctx context.Context, item *models.Item) error {
	err := r.db.WithContext(ctx).
		Omit("ID", "CreatedAt", "UpdatedAt").
		Create(item).Error
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}
	return nil
}

// UpdateItem updates an existing item
func (r *repository) UpdateItem(ctx context.Context, item *models.Item) error {
	// First check if record exists
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Item{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check item existence: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("item not found: %w", ErrNotFound)
	}

	// Update the record
	err := r.db.WithContext(ctx).
		Model(&models.Item{}).
		Where("id = ?", item.ID).
		Omit("ID", "CreatedAt").
		Updates(map[string]interface{}{
			"name":        item.Name,
			"description": item.Description,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}
	return nil
}

// DeleteItem deletes an item by ID
func (r *repository) DeleteItem(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Item{}, id)
	if err := checkRowsAffected(result); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}
