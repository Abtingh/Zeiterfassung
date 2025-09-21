// util/config.go
package util

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	APPPORT      string `mapstructure:"APP_PORT"`
	APPNAME      string `mapstructure:"APP_NAME"`
	APPDEBUG     string `mapstructure:"APP_DEBUG"`
	DBCONNECTION string `mapstructure:"DB_CONNECTION"`
	DBHOST       string `mapstructure:"DB_HOST"`
	DBUSERNAME   string `mapstructure:"DB_USERNAME"`
	DBPASSWORD   string `mapstructure:"DB_PASSWORD"`
	DBDATABASE   string `mapstructure:"DB_DATABASE"`
	DBPORT       string `mapstructure:"DB_PORT"`
	MIGRATIONURL string `mapstructure:"MIGRATION_URL"`
	JWT_SECRET   string `mapstructure:"JWT_SECRET"`
}

func LoadConfig() (config Config, err error) {
	// No files. Env only.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Bind the keys so Unmarshal includes env values
	keys := []string{
		"APP_PORT", "APP_NAME", "APP_DEBUG",
		"DB_CONNECTION", "DB_HOST", "DB_USERNAME", "DB_PASSWORD", "DB_DATABASE", "DB_PORT",
		"MIGRATION_URL", "JWT_SECRET",
	}
	for _, k := range keys {
		_ = viper.BindEnv(k)
	}

	// Optional sane defaults (won’t override env)
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_DEBUG", "false")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("MIGRATION_URL", "file://internal/db/migration")

	if err = viper.Unmarshal(&config); err != nil {
		return config, err
	}
	// Minimal validation
	if config.JWT_SECRET == "" {
		return config, errors.New("missing required env: JWT_SECRET")
	}
	if config.DBDATABASE == "" || config.DBHOST == "" {
		return config, errors.New("missing required DB_* envs")
	}
	return config, nil
}
