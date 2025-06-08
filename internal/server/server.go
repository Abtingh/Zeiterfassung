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
// NewServer creates and returns a new Server instance configured with the given address and database queries.
// It sets up an HTTP request multiplexer, serves static files from the "./public/" directory, and registers
// application routes using the provided queries. The returned Server is ready to be started to handle HTTP requests.
//
// Parameters:
//   - addr: The address the server will listen on (e.g., ":8080").
//   - queries: A pointer to a db.Queries instance for database operations.
//
// Returns:
//   - A pointer to the initialized Server.
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
