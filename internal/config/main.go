package config

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	AppName = "deviceregistry"
)

func ReadConfig(envFlag string, configDir string) {
	setDefaults(envFlag)

	viper.SetConfigName("config-" + envFlag)
	viper.AddConfigPath(fmt.Sprintf("/etc/%s/", AppName))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%s/", AppName))
	viper.AddConfigPath(".")
	if configDir != "" {
		viper.AddConfigPath(configDir)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("config file not found: %s", err)
	}

	viper.SetEnvPrefix(AppName)
	r := strings.NewReplacer(
		".", "_",
		"-", "_",
	)
	viper.SetEnvKeyReplacer(r)
	viper.AutomaticEnv()

	// Setup logging
	if viper.GetBool("log.structured") {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
	}

	log.SetLevel(log.Level(viper.GetUint32("log.level")))
}

func setDefaults(envFlag string) {
	// Server defaults
	viper.SetDefault("port", 8080)

	// Database defaults
	viper.SetDefault("db-max-open-connections", 25)
	viper.SetDefault("db-max-idle-time", 5)
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 5432)
	viper.SetDefault("db.user", "deviceregistry")
	viper.SetDefault("db.password", "deviceregistry123")
	viper.SetDefault("db.name", "deviceregistry")
	viper.SetDefault("db.ssl-mode", "disable")

	// Migration defaults
	switch envFlag {
	case "test":
		viper.SetDefault("migration.dir", "./db/migrations")
	default:
		viper.SetDefault("migration.dir", "./db/migrations")
	}

	// Cache defaults
	viper.SetDefault("cache.mute", true)
	viper.SetDefault("cache.verbose", false)
	viper.SetDefault("cache.ttl", 10*time.Minute)

	// Logging defaults
	viper.SetDefault("log.structured", false)
	viper.SetDefault("log.level", uint32(log.InfoLevel))
}
