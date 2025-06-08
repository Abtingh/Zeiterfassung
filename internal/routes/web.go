package routes

import (
	"net/http"

	"github.com/Abtingh/Zeiterfassung/internal/api/handlers"
)

func Register(mux *http.ServeMux, handler *handlers.Handler) {
	mux.HandleFunc("/login", handler.LoginHandler)
	mux.HandleFunc("/logout", handler.LogoutHandler)
	mux.HandleFunc("/home/", handler.HomeHandler)

	// Errors
	mux.HandleFunc("/error/authentication", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/public-static/error/authentication.html")
	})
}
