package handlers

import (
	"fmt"
	"log"
	"net/http"
)

type LoginRequest struct {
	Email    string `json:"email"`
	password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token,omitempty"`
	Erorr string `json:"error,omitempty"`
}

var jwtSecretKey []byte

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./public/public-static/login.html")
		return
	}

	if r.Method == http.MethodPost {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("ParseForm() error: %v", err), http.StatusBadRequest)
			return
		}

		// Extract form values
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Write response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "POST request successful\n")
		fmt.Fprintf(w, "Email: %s\n", email)
		fmt.Fprintf(w, "Password: %s\n", password)
		return
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {}
