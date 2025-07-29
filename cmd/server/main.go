package main

import (
	"subscription_tracker_api/internal/config"
	"subscription_tracker_api/internal/handlers"
	"subscription_tracker_api/internal/repository"
	"subscription_tracker_api/internal/service"

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
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.Info("Starting Subscription Tracker API...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration: ", err)
	}

	// Connect to database
	db, err := repository.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		logger.Fatal("Failed to run migrations: ", err)
	}

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
