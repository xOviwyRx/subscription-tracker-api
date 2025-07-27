package main

import (
	"log"
	"subscription_tracker_api/internal/config"
	"subscription_tracker_api/internal/repository"
)

func main() {
	log.Println("Testing database connection...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	log.Printf("Connecting to database: %s:%s/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	// Connect to database
	db, err := repository.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("Database connected successfully!")

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	log.Println("✅ Database setup completed successfully!")
}
