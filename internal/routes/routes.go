package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/GunarsK-templates/template-api/internal/config"
	"github.com/GunarsK-templates/template-api/internal/handlers"
	// Uncomment after running: swag init -g cmd/api/main.go -o docs
	// _ "github.com/GunarsK-templates/template-api/docs"
)

// Setup configures all routes for the service
func Setup(router *gin.Engine, handler *handlers.Handler, cfg *config.Config) {
	// CORS middleware
	router.Use(corsMiddleware(cfg.Service.AllowedOrigins))

	// Security headers
	router.Use(securityHeaders())

	// Health check (unprotected)
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no auth required)
		items := v1.Group("/items")
		{
			items.GET("", handler.GetItems)
			items.GET("/:id", handler.GetItem)
		}

		// Protected routes (require auth)
		// Uncomment and add auth middleware when JWT is configured
		// if cfg.HasJWT() {
		//     protected := v1.Group("")
		//     protected.Use(authMiddleware(cfg.JWT.Secret))
		//     {
		//         protected.POST("/items", handler.CreateItem)
		//         protected.PUT("/items/:id", handler.UpdateItem)
		//         protected.DELETE("/items/:id", handler.DeleteItem)
		//     }
		// }

		// For now, all routes are public (remove in production)
		items.POST("", handler.CreateItem)
		items.PUT("/:id", handler.UpdateItem)
		items.DELETE("/:id", handler.DeleteItem)
	}

	// Swagger documentation (only if host is configured)
	if cfg.Service.SwaggerHost != "" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

// corsMiddleware handles CORS preflight and headers
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "86400")
		}

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// securityHeaders adds security headers to all responses
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
