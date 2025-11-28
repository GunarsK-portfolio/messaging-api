package repository

import (
	"context"

	"github.com/GunarsK-portfolio/portfolio-common/models"
	commonrepo "github.com/GunarsK-portfolio/portfolio-common/repository"
	"gorm.io/gorm"
)

// Repository defines the interface for messaging data operations
type Repository interface {
	// Contact Messages (public: create only, admin: list/get)
	CreateContactMessage(ctx context.Context, message *models.ContactMessage) error
	GetContactMessages(ctx context.Context) ([]models.ContactMessage, error)
	GetContactMessageByID(ctx context.Context, id int64) (*models.ContactMessage, error)
	UpdateContactMessageStatus(ctx context.Context, id int64, status string, lastError *string) error

	// Recipients (admin only)
	GetAllRecipients(ctx context.Context) ([]models.Recipient, error)
	GetActiveRecipients(ctx context.Context) ([]models.Recipient, error)
	GetRecipientByID(ctx context.Context, id int64) (*models.Recipient, error)
	CreateRecipient(ctx context.Context, recipient *models.Recipient) error
	UpdateRecipient(ctx context.Context, recipient *models.Recipient) error
	DeleteRecipient(ctx context.Context, id int64) error
}

type repository struct {
	db *gorm.DB
	*commonrepo.SafeUpdater
}

// New creates a new repository instance
func New(db *gorm.DB) Repository {
	return &repository{
		db:          db,
		SafeUpdater: commonrepo.NewSafeUpdater(db),
	}
}

// checkRowsAffected wraps the common helper
var checkRowsAffected = commonrepo.CheckRowsAffected

// safeUpdate wraps SafeUpdater.Update for backward compatibility
func (r *repository) safeUpdate(ctx context.Context, model interface{}, id int64) error {
	return r.Update(ctx, model, id)
}
