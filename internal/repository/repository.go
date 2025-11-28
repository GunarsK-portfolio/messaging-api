package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/GunarsK-templates/template-api/internal/config"
	"github.com/GunarsK-templates/template-api/internal/models"
)

// Repository defines the interface for data access
type Repository interface {
	// Item operations
	GetAllItems(ctx context.Context) ([]models.Item, error)
	GetItemByID(ctx context.Context, id int64) (*models.Item, error)
	CreateItem(ctx context.Context, item *models.Item) error
	UpdateItem(ctx context.Context, item *models.Item) error
	DeleteItem(ctx context.Context, id int64) error
}

type repository struct {
	db *gorm.DB
}

// New creates a new repository instance
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ConnectDB establishes a database connection
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	// Configure GORM logger
	gormLogLevel := logger.Silent
	if cfg.Service.Environment == "development" {
		gormLogLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	slog.Info("Database connection established")

	return db, nil
}

// checkRowsAffected verifies that at least one row was affected
func checkRowsAffected(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
