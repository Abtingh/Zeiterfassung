package main

import (
	"fmt"
	"log"

	api "github.com/Abtingh/Zeiterfassung/internal/api/handlers"
	"github.com/Abtingh/Zeiterfassung/internal/util"
)

func main() {

	// Konfiguration laden
	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("Konfiguration konnte nicht geladen werden: %v", err)
	}

	log.Printf("Debug-Modus: %s\n", cfg.APPDEBUG)
	log.Printf("Lausche auf Port: %s\n", cfg.APPPORT)

	// Beispiel für den Aufbau einer Datenbank-Verbindungszeichenfolge (DSN)
	// Die Werte werden aus der Konfiguration geladen
	dsn := fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBCONNECTION, // z.B. "postgres"
		cfg.DBUSERNAME,   // z.B. "root"
		cfg.DBPASSWORD,   // z.B. "secret"
		cfg.DBHOST,       // z.B. "localhost"
		cfg.DBPORT,       // z.B. "5432"
		cfg.DBDATABASE,   // z.B. "zeiterfassung"
	)
	log.Printf("DSN: %s\n", dsn)

	// Server-Adresse erstellen und neuen Server starten
	address := fmt.Sprintf(":%s", cfg.APPPORT)
	server := api.NewServer(address)

	// Server starten und Fehler behandeln
	if err := server.Start(); err != nil {
		log.Fatalf("Server-Fehler: %v", err)
	}
}
