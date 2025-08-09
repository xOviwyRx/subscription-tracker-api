package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"subscription_tracker_api/internal/config"
	"subscription_tracker_api/internal/handlers"
	"subscription_tracker_api/internal/repository"
	"subscription_tracker_api/internal/service"
	"syscall"
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
	log.Println("Loading application configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration: ", err)
	}

	// Initialize logger with configuration
	logger := setupLogger(cfg.Logging)
	logger.Info("Starting Subscription Tracker API...")
	logger.WithFields(logrus.Fields{
		"host": cfg.Server.Host,
		"port": cfg.Server.Port,
	}).Info("Configuration loaded successfully")

	// Connect to database
	logger.Info("Establishing database connection...")
	db, err := repository.NewDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database: ", err)
	}
	logger.Info("Database connection established successfully")

	// Run migrations
	logger.Info("Running database migrations...")
	if err := db.RunMigrations(); err != nil {
		logger.WithError(err).Fatal("Migration error")
	}
	logger.Info("Database migrations completed successfully")

	// Initialize repository
	logger.Info("Initializing repository layer...")
	subscriptionRepo := repository.NewSubscriptionRepository(db.DB, logger)
	logger.Info("Repository layer initialized successfully")

	// Initialize service
	logger.Info("Initializing service layer...")
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, logger)
	logger.Info("Service layer initialized successfully")

	// Initialize handlers
	logger.Info("Initializing HTTP handlers...")
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService, logger)
	logger.Info("HTTP handlers initialized successfully")

	// Setup Gin router
	logger.Info("Setting up HTTP router and middleware...")
	router := gin.Default()

	// Add request logging middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.WithFields(logrus.Fields{
			"method":     param.Method,
			"path":       param.Path,
			"status":     param.StatusCode,
			"latency":    param.Latency,
			"client_ip":  param.ClientIP,
			"user_agent": param.Request.UserAgent(),
		}).Info("HTTP request processed")
		return ""
	}))

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
	logger.Info("CORS middleware configured successfully")

	// API routes
	logger.Info("Configuring API routes...")
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
	logger.WithField("routes_count", 6).Info("API routes configured successfully")

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		logger.Debug("Health check endpoint accessed")
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "subscription-tracker-api",
		})
	})
	logger.Info("Health check endpoint configured")

	// Swagger documentation
	logger.Info("Configuring Swagger documentation...")
	router.Static("/docs", "./docs")
	url := ginSwagger.URL("http://localhost:8080/docs/swagger.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	logger.Info("Swagger documentation configured at /swagger/index.html")

	// Create HTTP server
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logger.WithFields(logrus.Fields{
			"address": serverAddr,
			"version": "1.0",
		}).Info("Starting HTTP server...")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	logger.Info("Server started successfully. Press Ctrl+C to gracefully shutdown...")

	// Wait for interrupt signal
	<-quit
	logger.Info("Shutdown signal received, initiating graceful shutdown...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server
	logger.Info("Shutting down HTTP server...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("HTTP server shutdown gracefully")
	}

	// Close database connection
	logger.Info("Closing database connection...")
	db.Close()
	logger.Info("Database connection closed")

	logger.Info("Application shutdown completed")
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
	}).Info("Logger initialized successfully")

	return logger
}
