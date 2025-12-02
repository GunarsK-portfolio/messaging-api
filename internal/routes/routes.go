package routes

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/GunarsK-portfolio/messaging-api/docs"
	"github.com/GunarsK-portfolio/messaging-api/internal/config"
	"github.com/GunarsK-portfolio/messaging-api/internal/handlers"
	"github.com/GunarsK-portfolio/portfolio-common/health"
	"github.com/GunarsK-portfolio/portfolio-common/jwt"
	"github.com/GunarsK-portfolio/portfolio-common/metrics"
	common "github.com/GunarsK-portfolio/portfolio-common/middleware"
)

// Setup configures all routes for the service
func Setup(router *gin.Engine, handler *handlers.Handler, cfg *config.Config, metricsCollector *metrics.Metrics, healthAgg *health.Aggregator) {
	// Security middleware with CORS validation
	securityMiddleware := common.NewSecurityMiddleware(
		cfg.AllowedOrigins,
		"GET,POST,PUT,DELETE,OPTIONS",
		"Content-Type,Authorization",
		true,
	)
	router.Use(securityMiddleware.Apply())

	// Health check
	router.GET("/health", healthAgg.Handler())

	// Metrics
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Public routes (no auth required)
	public := v1.Group("/contact")
	{
		public.POST("", handler.CreateContactMessage)
	}

	// Protected routes (require JWT auth)
	jwtService, err := jwt.NewValidatorOnly(cfg.JWTSecret)
	if err != nil {
		log.Fatalf("Failed to create JWT service: %v", err)
	}

	authMiddleware := common.NewAuthMiddleware(jwtService)

	protected := v1.Group("")
	protected.Use(authMiddleware.ValidateToken())
	protected.Use(authMiddleware.AddTTLHeader())
	{
		// Contact messages (read-only for admin)
		messages := protected.Group("/messages")
		{
			messages.GET("", handler.GetContactMessages)
			messages.GET("/:id", handler.GetContactMessage)
		}

		// Recipients management (full CRUD for admin)
		recipients := protected.Group("/recipients")
		{
			recipients.GET("", handler.GetRecipients)
			recipients.GET("/:id", handler.GetRecipient)
			recipients.POST("", handler.CreateRecipient)
			recipients.PUT("/:id", handler.UpdateRecipient)
			recipients.DELETE("/:id", handler.DeleteRecipient)
		}
	}

	// Swagger documentation (only if host is configured)
	if cfg.SwaggerHost != "" {
		docs.SwaggerInfo.Host = cfg.SwaggerHost
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
