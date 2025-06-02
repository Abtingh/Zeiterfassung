package util

import "github.com/spf13/viper"

// Config repräsentiert die Konfigurationsstruktur für die Anwendung.
// Sie enthält alle notwendigen Einstellungen für die App und Datenbankverbindung.
type Config struct {
	APPPORT      string `mapstructure:"APP_PORT"`      // Port der Anwendung
	APPNAME      string `mapstructure:"APP_NAME"`      // Name der Anwendung
	APPDEBUG     string `mapstructure:"APP_DEBUG"`     // Debug-Modus-Einstellung
	DBCONNECTION string `mapstructure:"DB_CONNECTION"` // Art der Datenbankverbindung
	DBHOST       string `mapstructure:"DB_HOST"`       // Datenbank-Host
	DBUSERNAME   string `mapstructure:"DB_USERNAME"`   // Datenbank-Benutzername
	DBPASSWORD   string `mapstructure:"DB_PASSWORD"`   // Datenbank-Passwort
	DBDATABASE   string `mapstructure:"DB_DATABASE"`   // Name der Datenbank
	DBPORT       string `mapstructure:"DB_PORT"`       // Port der Datenbank
	MIGRATIONURL string `mapstructure:"MIGRATION_URL"` // URL für Migrationsdateien
	JWT_SECRET   string `mapstructure:"JWT_SECRET"`
}

// LoadConfig lädt die Konfigurationsdaten aus einer .env Datei am angegebenen Pfad.
// Es gibt die geladene Konfiguration und einen möglichen Fehler zurück.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)  // Setzt den Pfad zur Konfigurationsdatei
	viper.SetConfigType("env") // Setzt den Typ der Konfigurationsdatei auf .env
	viper.SetConfigName("app") // Setzt den Namen der Konfigurationsdatei auf "app"

	viper.AutomaticEnv() // Ermöglicht das Überschreiben durch Umgebungsvariablen

	// Liest die Konfigurationsdatei ein
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	// Wandelt die gelesenen Werte in die Config-Struktur um
	err = viper.Unmarshal(&config)
	return
}
