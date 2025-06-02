package util

import "github.com/spf13/viper"

type Config struct {
	APPPORT  string `mapstructure:"APP_PORT"`
	APPNAME  string `mapstructure:"APP_NAME"`
	APPDEBUG string `mapstructure:"APP_DEBUG"`
	DBCONNECTION string `mapstructure:"DB_CONNECTION"`
	DBHOST string `mapstructure:"DB_HOST"`
	DBUSERNAME string `mapstructure:"DB_USERNAME"`
	DBPASSWORD string `mapstructure:"DB_PASSWORD"`
	DBDATABASE string `mapstructure:"DB_DATABSE"`
	DBPORT string `mapstructure:"DB_PORT"`
	MIGRATIONURL string `mapstructure:"MIGRATION_URL"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("app")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return

}
