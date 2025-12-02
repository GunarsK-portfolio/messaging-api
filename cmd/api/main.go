package main

import (
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	_ "github.com/GunarsK-portfolio/messaging-api/docs"
	"github.com/GunarsK-portfolio/messaging-api/internal/config"
	"github.com/GunarsK-portfolio/messaging-api/internal/handlers"
	"github.com/GunarsK-portfolio/messaging-api/internal/repository"
	"github.com/GunarsK-portfolio/messaging-api/internal/routes"
	commondb "github.com/GunarsK-portfolio/portfolio-common/database"
	"github.com/GunarsK-portfolio/portfolio-common/logger"
	"github.com/GunarsK-portfolio/portfolio-common/metrics"
	"github.com/GunarsK-portfolio/portfolio-common/queue"
	"github.com/GunarsK-portfolio/portfolio-common/server"
)

// @title           Messaging API
// @version         1.0
// @description     API for contact form submissions and recipient management
// @host            localhost:8086
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.Load()

	appLogger := logger.New(logger.Config{
		Level:       os.Getenv("LOG_LEVEL"),
		Format:      os.Getenv("LOG_FORMAT"),
		ServiceName: "messaging-api",
		AddSource:   os.Getenv("LOG_SOURCE") == "true",
	})

	appLogger.Info("Starting messaging API", "version", "1.0")

	metricsCollector := metrics.New(metrics.Config{
		ServiceName: "messaging",
		Namespace:   "portfolio",
	})

	//nolint:staticcheck // Embedded field name required due to ambiguous fields
	db, err := commondb.Connect(commondb.PostgresConfig{
		Host:     cfg.DatabaseConfig.Host,
		Port:     strconv.Itoa(cfg.DatabaseConfig.Port),
		User:     cfg.DatabaseConfig.User,
		Password: cfg.DatabaseConfig.Password,
		DBName:   cfg.DatabaseConfig.Name,
		SSLMode:  cfg.DatabaseConfig.SSLMode,
		TimeZone: "UTC",
	})
	if err != nil {
		appLogger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	appLogger.Info("Database connection established")

	publisher, err := queue.NewRabbitMQPublisher(cfg.RabbitMQConfig)
	if err != nil {
		appLogger.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()
	appLogger.Info("RabbitMQ connection established")

	repo := repository.New(db)
	handler := handlers.New(repo, publisher)

	router := gin.New()
	router.Use(logger.Recovery(appLogger))
	router.Use(logger.RequestLogger(appLogger))
	router.Use(metricsCollector.Middleware())

	routes.Setup(router, handler, cfg, metricsCollector)

	appLogger.Info("Messaging API ready", "port", cfg.ServiceConfig.Port, "environment", os.Getenv("ENVIRONMENT"))

	serverCfg := server.DefaultConfig(strconv.Itoa(cfg.ServiceConfig.Port))
	if err := server.Run(router, serverCfg, appLogger); err != nil {
		appLogger.Error("Server error", "error", err)
		os.Exit(1)
	}
}
