package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/GunarsK-templates/template-api/internal/utils"
)

// ServiceConfig holds service-level configuration (port, environment, CORS)
type ServiceConfig struct {
	Name           string   `validate:"required"`
	Port           string   `validate:"required"`
	Environment    string   `validate:"required,oneof=development staging production"`
	AllowedOrigins []string `validate:"required,min=1"`
	SwaggerHost    string   // Optional: Swagger UI host. Empty disables swagger.
}

// NewServiceConfig loads service configuration from environment variables
func NewServiceConfig() ServiceConfig {
	// Parse allowed origins from comma-separated string
	allowedOriginsStr := utils.GetEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	rawOrigins := strings.Split(allowedOriginsStr, ",")
	allowedOrigins := make([]string, 0, len(rawOrigins))
	for _, origin := range rawOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowedOrigins = append(allowedOrigins, trimmed)
		}
	}

	cfg := ServiceConfig{
		Name:           utils.GetEnv("SERVICE_NAME", "your-service"),
		Port:           utils.GetEnv("PORT", "8080"),
		Environment:    utils.GetEnv("ENVIRONMENT", "development"),
		AllowedOrigins: allowedOrigins,
		SwaggerHost:    utils.GetEnv("SWAGGER_HOST", ""),
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		panic(fmt.Sprintf("Invalid service configuration: %v", err))
	}

	return cfg
}
