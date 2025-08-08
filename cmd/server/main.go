package main

import (
	"log"
	"subscription_tracker_api/internal/config"
	"subscription_tracker_api/internal/handlers"
	"subscription_tracker_api/internal/repository"
	"subscription_tracker_api/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Subscription Tracker API
// @version 1.0
// @description REST API for managing user subscriptions
// @host localhost:8080
// @BasePath /api/v1
func main() {

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration: ", err)
	}

	// Initialize logger with configuration
	logger := setupLogger(cfg.Logging)
	logger.Info("Starting Subscription Tracker API...")

	// Connect to database
	db, err := repository.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	log.Println("Migrations applied successfully")

	// Initialize repository
	subscriptionRepo := repository.NewSubscriptionRepository(db.DB)

	// Initialize service
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, logger)

	// Initialize handlers
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService, logger)

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// CRUDL operations for subscriptions
		v1.POST("/subscriptions", subscriptionHandler.CreateSubscription)
		v1.GET("/subscriptions/:id", subscriptionHandler.GetSubscription)
		v1.PUT("/subscriptions/:id", subscriptionHandler.UpdateSubscription)
		v1.DELETE("/subscriptions/:id", subscriptionHandler.DeleteSubscription)
		v1.GET("/subscriptions", subscriptionHandler.ListSubscriptions)

		// Cost calculation endpoint
		v1.GET("/subscriptions/cost", subscriptionHandler.CalculateTotalCost)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "subscription-tracker-api",
		})
	})
	// Swagger documentation
	router.Static("/docs", "./docs")
	url := ginSwagger.URL("http://localhost:8080/docs/swagger.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Start server
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	logger.WithField("address", serverAddr).Info("Server starting...")

	if err := router.Run(serverAddr); err != nil {
		logger.Fatal("Failed to start server: ", err)
	}
}

func setupLogger(loggingConfig config.LoggingConfig) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(loggingConfig.Level)
	if err != nil {
		level = logrus.InfoLevel
		logger.Warnf("Invalid log level '%s', defaulting to info", loggingConfig.Level)
	}
	logger.SetLevel(level)

	// Set formatter
	switch loggingConfig.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.Warnf("Invalid log format '%s', defaulting to json", loggingConfig.Format)
	}

	logger.WithFields(logrus.Fields{
		"level":  loggingConfig.Level,
		"format": loggingConfig.Format,
	}).Info("Logger initialized")

	return logger
}
