package server

import (
	"net/http"

	"github.com/Abtingh/Zeiterfassung/internal/api/handlers"
	"github.com/Abtingh/Zeiterfassung/internal/routes"
	db "github.com/Abtingh/Zeiterfassung/internal/db/sqlc"
)

// Server repräsentiert den HTTP-Server mit Adresse und Mux.
type Server struct {
	Addr string
	Mux  *http.ServeMux
}

// NewServer erstellt eine neue Server-Instanz mit gegebener Adresse.
func NewServer(addr string, queries *db.Queries) *Server {
	mux := http.NewServeMux()

	// Stellt statische Dateien aus dem Verzeichnis bereit.
	mux.Handle("/", http.FileServer(http.Dir("./public/")))

	// Registriert die Routen aus dem Package routes.
	handler := handlers.NewHandler(queries)
	routes.Register(mux, handler)

	return &Server{
		Addr: addr,
		Mux:  mux,
	}
}

// Start startet den HTTP-Server und gibt ggf. einen Fehler zurück.
func (s *Server) Start() error {
	return http.ListenAndServe(s.Addr, s.Mux)
}
