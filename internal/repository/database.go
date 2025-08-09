package repository

import (
	"fmt"
	"subscription_tracker_api/internal/config"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/golang-migrate/migrate/v4"
	migrate_pg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Database holds the database connection
type Database struct {
	DB     *gorm.DB
	logger *logrus.Logger
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config, logger *logrus.Logger) (*Database, error) {
	logger.Info("Initializing database connection...")

	dsn := cfg.GetDatabaseDSN()
	logger.WithFields(logrus.Fields{
		"host":     cfg.Database.Host,
		"port":     cfg.Database.Port,
		"database": cfg.Database.DBName,
		"user":     cfg.Database.User,
	}).Info("Connecting to database with configuration")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		logger.WithError(err).Error("Failed to establish database connection")
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Info("GORM database connection established")

	// Test the connection
	logger.Info("Testing database connection...")
	sqlDB, err := db.DB()
	if err != nil {
		logger.WithError(err).Error("Failed to get underlying sql.DB instance")
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		logger.WithError(err).Error("Database ping failed")
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	logger.Info("Database ping successful - connection verified")

	// Configure connection pool
	logger.Info("Configuring database connection pool...")
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	logger.WithFields(logrus.Fields{
		"max_idle_connections": 10,
		"max_open_connections": 100,
	}).Info("Database connection pool configured")

	return &Database{
		DB:     db,
		logger: logger,
	}, nil
}

// RunMigrations runs the database migrations
func (d *Database) RunMigrations() error {
	d.logger.Info("Starting database migration process...")

	sqlDB, err := d.DB.DB()
	if err != nil {
		d.logger.WithError(err).Error("Failed to get sql.DB instance for migrations")
		return fmt.Errorf("failed to get sql.DB from GORM: %w", err)
	}

	d.logger.Info("Creating migration driver...")
	driver, err := migrate_pg.WithInstance(sqlDB, &migrate_pg.Config{})
	if err != nil {
		d.logger.WithError(err).Error("Failed to create migration driver")
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	d.logger.Info("Initializing migration instance...")
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres", driver)
	if err != nil {
		d.logger.WithError(err).Error("Failed to create migration instance")
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	d.logger.Info("Executing database migrations...")
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		d.logger.WithError(err).Error("Migration execution failed")
		return fmt.Errorf("migration error: %w", err)
	}

	if err == migrate.ErrNoChange {
		d.logger.Info("No new migrations to apply - database schema is up to date")
	} else {
		d.logger.Info("All database migrations executed successfully")
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	d.logger.Info("Closing database connection...")

	sqlDB, err := d.DB.DB()
	if err != nil {
		d.logger.WithError(err).Error("Failed to get sql.DB instance for closing")
		return err
	}

	err = sqlDB.Close()
	if err != nil {
		d.logger.WithError(err).Error("Failed to close database connection")
	} else {
		d.logger.Info("Database connection closed successfully")
	}

	return err
}
