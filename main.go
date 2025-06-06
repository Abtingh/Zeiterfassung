package main

import (
	"context"
	"fmt"
	"log"
	"time"

	db "github.com/Abtingh/Zeiterfassung/internal/db/sqlc"
	"github.com/Abtingh/Zeiterfassung/internal/server"
	"github.com/Abtingh/Zeiterfassung/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
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

	// DatenBnk connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Unable to parse DATABSE_URL: %v", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Unable to create a connection pool: %v", err)
	}

	defer dbPool.Close()

	queries := db.New(dbPool)

	// Server-Adresse erstellen und neuen Server starten
	address := fmt.Sprintf(":%s", cfg.APPPORT)
	server := server.NewServer(address, queries)

	// Server starten und Fehler behandeln
	if err := server.Start(); err != nil {
		log.Fatalf("Server-Fehler: %v", err)
	}
}
