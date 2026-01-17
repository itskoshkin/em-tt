package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

const (
	LogEnabled   = "app.log.enabled"
	LogToFile    = "app.log.file_path"
	LogKeepFiles = "app.log.keep_files"

	ApiHost        = "app.api.host"
	ApiPort        = "app.api.port"
	ApiBasePath    = "app.api.base_path"
	GinReleaseMode = "app.api.gin_release_mode"

	DatabaseHost     = "app.database.host"
	DatabasePort     = "app.database.port"
	DatabaseUser     = "app.database.user"
	DatabasePassword = "app.database.password"
	DatabaseName     = "app.database.database_name"
	DatabaseSslMode  = "app.database.ssl_mode"
)

func LoadConfig() {
	fmt.Print("Loading configuration... ")

	viper.SetConfigFile("./config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println()
		log.Fatalf("Fatal: failed to read configuration: %v", err)
	}

	fmt.Println(" Done.")
}
