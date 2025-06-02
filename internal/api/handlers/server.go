package handlers

import (
	"log"
	"net/http"

	"github.com/Abtingh/Zeiterfassung/internal/routes"
)

// Server repräsentiert den HTTP-Server mit Adresse und Mux.
type Server struct {
	Addr string
	Mux  *http.ServeMux
}

// NewServer erstellt eine neue Server-Instanz mit gegebener Adresse.
func NewServer(addr string) *Server {
	mux := http.NewServeMux()

	// Stellt statische Dateien aus dem Verzeichnis bereit.
	mux.Handle("/", http.FileServer(http.Dir("./public/public-static")))

	// Registriert die Routen aus dem Package routes.
	routes.Register(mux)

	return &Server{
		Addr: addr,
		Mux:  mux,
	}
}

// Start startet den HTTP-Server und gibt ggf. einen Fehler zurück.
func (s *Server) Start() error {
	log.Printf("Starting server on %s\n", s.Addr)
	return http.ListenAndServe(s.Addr, s.Mux)
}
