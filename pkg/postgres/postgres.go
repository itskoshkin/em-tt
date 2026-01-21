package postgres

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
	LogLevel string
}

func NewInstance(cfg Config) *gorm.DB {
	fmt.Print("Connecting to Postgres... ")

	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)), &gorm.Config{})
	if err != nil {
		fmt.Println()
		log.Fatalf("Fatal: failed to connect to database: %v", err)
	}

	switch cfg.LogLevel {
	case "DEBUG", "INFO":
		db.Config.Logger.LogMode(logger.Info)
	case "WARN":
		db.Config.Logger.LogMode(logger.Warn)
	case "ERROR":
		db.Config.Logger.LogMode(logger.Error)
	}

	fmt.Println("Done.")
	return db
}
