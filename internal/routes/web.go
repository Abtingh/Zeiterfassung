package routes

import (
	"net/http"

	api "github.com/Abtingh/Zeiterfassung/internal/api/handlers"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("/login", api.LoginHandler)
	mux.HandleFunc("/register", api.RegisterHandler)
	mux.HandleFunc("/login", api.LogoutHandler)
}
