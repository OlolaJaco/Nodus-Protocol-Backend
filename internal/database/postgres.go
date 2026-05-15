package database

import (
	"fmt"
	"time"

	"github.com/nodus-protocol/backend/internal/config"
	"github.com/nodus-protocol/backend/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewPostgres creates and returns a new GORM database connection.
func NewPostgres(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	logLevel := gormlogger.Silent
	if !cfg.IsProd() {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(logLevel),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("PostgreSQL connected", zap.String("host", cfg.Database.Host), zap.String("db", cfg.Database.Name))
	return db, nil
}

// AutoMigrate runs GORM auto-migrations for all models.
// Use golang-migrate for production schema changes; this is for dev convenience.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Token{},
	)
}
