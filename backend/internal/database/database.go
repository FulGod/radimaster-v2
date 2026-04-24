package database

import (
	"log"

	"radimaster/internal/config"
	"radimaster/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a database connection. Uses PostgreSQL if DATABASE_URL is set, otherwise SQLite.
func Connect(cfg *config.Config) *gorm.DB {
	var logLevel logger.LogLevel
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	} else {
		logLevel = logger.Warn
	}

	gormCfg := &gorm.Config{Logger: logger.Default.LogMode(logLevel)}

	var db *gorm.DB
	var err error

	if cfg.DatabaseURL != "" {
		db, err = gorm.Open(postgres.Open(cfg.DatabaseURL), gormCfg)
		log.Println("🐘 Connected to PostgreSQL")
	} else {
		db, err = gorm.Open(sqlite.Open("radimaster.db"), gormCfg)
		log.Println("📦 Using SQLite (development mode)")
	}
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	if cfg.DatabaseURL != "" {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(10)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.MedicalCase{},
		&models.Phase{},
		&models.Slice{},
		&models.Classroom{},
	); err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}

	log.Println("✅ Database connected and migrated")
	return db
}
