package routes

import (
	"net/http"

	"github.com/Abtingh/Zeiterfassung/internal/api/handlers"
)

func Register(mux *http.ServeMux, handler *handlers.Handler) {
	mux.HandleFunc("/login", handler.LoginHandler)
	mux.HandleFunc("/logout", handler.LogoutHandler)
}
