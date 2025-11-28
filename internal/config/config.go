package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	common "github.com/GunarsK-portfolio/portfolio-common/config"
)

// Config holds all configuration for the messaging service
type Config struct {
	common.DatabaseConfig
	common.ServiceConfig
	JWTSecret string `validate:"required,min=32"`
}

// Load loads all configuration from environment variables
func Load() *Config {
	cfg := &Config{
		DatabaseConfig: common.NewDatabaseConfig(),
		ServiceConfig:  common.NewServiceConfig("8086"),
		JWTSecret:      common.GetEnvRequired("JWT_SECRET"),
	}

	// Validate service-specific fields
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		panic(fmt.Sprintf("Invalid configuration: %v", err))
	}

	return cfg
}
