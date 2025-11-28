package config

// Config holds all configuration for the service
type Config struct {
	Service  ServiceConfig
	Database DatabaseConfig
	JWT      *JWTConfig // Optional - nil if JWT_SECRET not set
}

// Load loads all configuration from environment variables
func Load() *Config {
	return &Config{
		Service:  NewServiceConfig(),
		Database: NewDatabaseConfig(),
		JWT:      NewJWTConfig(),
	}
}

// HasJWT returns true if JWT authentication is configured
func (c *Config) HasJWT() bool {
	return c.JWT != nil && c.JWT.HasJWT()
}
