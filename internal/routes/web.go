package routes

import (
	"net/http"

	"github.com/Abtingh/Zeiterfassung/internal/api/handlers"
	"github.com/Abtingh/Zeiterfassung/internal/api/middleware"
)

// Register sets up the HTTP routes for the application using the provided ServeMux and Handler.
// It registers handlers for login, logout, and home endpoints, as well as a custom error page
// for authentication errors.
//
// Parameters:
//
//	mux     - the HTTP request multiplexer to register routes on
//	handler - the Handler struct containing the handler functions for each route
func Register(mux *http.ServeMux, handler *handlers.Handler) {
	mux.HandleFunc("/login", middleware.NoCache(handler.LoginHandler))
	mux.HandleFunc("/logout", middleware.NoCache(handler.LogoutHandler))
	mux.HandleFunc("/home", middleware.NoCache(handler.HomeHandler))
	mux.HandleFunc("/reset-passwort/", middleware.NoCache(handler.ResetPasswordHandler))
	mux.HandleFunc("/me", middleware.NoCache(handler.MeHandler))

	// Student pages
	mux.HandleFunc("/ZeitEintragen", middleware.NoCache(handler.StudentZeitEintragenHandler))

	// Admin pages
	mux.HandleFunc("/admin/users/create", middleware.NoCache(handler.CreateUserHandler)) // POST
	mux.HandleFunc("/admin/users/list", middleware.NoCache(handler.ListUsersHandler))    // GET

	// Time Entry API endpoints
	mux.HandleFunc("/api/time-entries/submit", middleware.NoCache(handler.SubmitWeeklyTimeEntriesHandler))
	mux.HandleFunc("/api/time-entries/week", middleware.NoCache(handler.GetWeeklyTimeEntriesHandler))

	// Supervisor API endpoints
	mux.HandleFunc("/api/supervisor/pending-entries", middleware.NoCache(handler.GetPendingTimeEntriesHandler))

	// Errors
	mux.HandleFunc("/error/authentication", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/public-static/error/authentication.html")
	})
}
