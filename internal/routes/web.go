package routes

import (
	"net/http"

	"github.com/Abtingh/Zeiterfassung/internal/api/handlers"
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
	mux.HandleFunc("/login", handler.LoginHandler)
	mux.HandleFunc("/logout", handler.LogoutHandler)
	mux.HandleFunc("/home", handler.HomeHandler)
	mux.HandleFunc("/reset-passwort/", handler.ResetPasswordHandler)
	mux.HandleFunc("/me", handler.MeHandler)

	// Student pages
	mux.HandleFunc("/ZeitEintragen", handler.StudentZeitEintragenHandler)

	// Errors
	mux.HandleFunc("/error/authentication", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/public-static/error/authentication.html")
	})
}
