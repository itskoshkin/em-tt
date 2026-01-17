package postgres

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"subscription-aggregator-service/internal/config"
	"subscription-aggregator-service/internal/models"
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

	//db.Config.Logger.LogMode(logger.Error) //TODO: Get from config

	if err = db.AutoMigrate(&models.Subscription{}); err != nil {
		fmt.Println()
		log.Fatalf("Fatal: failed to migrate database: %v", err) //TODO: Text
	} //TODO: Remove when goosed?

	fmt.Println("Done.")
	return db
}
