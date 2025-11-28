package config

import (
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/GunarsK-templates/template-api/internal/utils"
)

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret        string        `validate:"required,min=32"`
	AccessExpiry  time.Duration `validate:"gt=0"`
	RefreshExpiry time.Duration `validate:"gt=0"`
}

// NewJWTConfig loads JWT configuration from environment variables.
// Returns nil if JWT_SECRET is not set (JWT is optional).
// Default values:
//   - JWT_ACCESS_EXPIRY: 15m (15 minutes)
//   - JWT_REFRESH_EXPIRY: 168h (7 days)
func NewJWTConfig() *JWTConfig {
	// JWT is optional - return nil if not configured
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil
	}

	cfg := &JWTConfig{
		Secret:        secret,
		AccessExpiry:  utils.GetEnvDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
		RefreshExpiry: utils.GetEnvDuration("JWT_REFRESH_EXPIRY", 168*time.Hour),
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		panic(fmt.Sprintf("Invalid JWT configuration: %v", err))
	}

	return cfg
}

// HasJWT returns true if JWT authentication is configured
func (c *JWTConfig) HasJWT() bool {
	return c != nil && c.Secret != ""
}
