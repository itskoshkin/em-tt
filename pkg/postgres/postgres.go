package postgres

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"subscription-aggregator-service/internal/config"
)

func NewInstance() *gorm.DB {
	fmt.Print("Connecting to Postgres... ")

	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString(config.DatabaseHost),
		viper.GetString(config.DatabasePort),
		viper.GetString(config.DatabaseUser),
		viper.GetString(config.DatabasePassword),
		viper.GetString(config.DatabaseName),
		viper.GetString(config.DatabaseSslMode),
	)), &gorm.Config{})
	if err != nil {
		fmt.Println()
		log.Fatalf("Fatal: failed to connect to database: %v", err)
	}

	switch viper.GetString(config.LogLevel) {
	case "DEBUG":
		db.Config.Logger.LogMode(logger.Info)
	case "INFO":
		db.Config.Logger.LogMode(logger.Info)
	case "WARN":
		db.Config.Logger.LogMode(logger.Warn)
	case "ERROR":
		db.Config.Logger.LogMode(logger.Error)
	}

	fmt.Println("Done.")
	return db
}
