package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/GunarsK-templates/template-api/internal/utils"
)

// DatabaseConfig holds PostgreSQL database configuration
type DatabaseConfig struct {
	Host     string `validate:"required"`
	Port     string `validate:"required"`
	User     string `validate:"required"`
	Password string `validate:"required"`
	Name     string `validate:"required"`
	SSLMode  string `validate:"required,oneof=disable require verify-ca verify-full"`
}

// NewDatabaseConfig loads database configuration from environment variables
func NewDatabaseConfig() DatabaseConfig {
	cfg := DatabaseConfig{
		Host:     utils.GetEnvRequired("DB_HOST"),
		Port:     utils.GetEnv("DB_PORT", "5432"),
		User:     utils.GetEnvRequired("DB_USER"),
		Password: utils.GetEnvRequired("DB_PASSWORD"),
		Name:     utils.GetEnvRequired("DB_NAME"),
		SSLMode:  utils.GetEnv("DB_SSL_MODE", "disable"),
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		panic(fmt.Sprintf("Invalid database configuration: %v", err))
	}

	return cfg
}

// DSN returns the database connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}
