package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Abtingh/Zeiterfassung/internal/util"
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

		user, ok := users[email]
		if !ok || !util.CheckPasswordHash(password, user.HashedPassword) {
			er := http.StatusUnauthorized
			http.Error(w, "Invalid username or password", er)
			return
		}
		sessionToken := util.GenerateToken(32)
		csrfToken := util.GenerateToken(32)


		//Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:	"session_token",
			Value:	sessionToken,
			Expires: time.Now().Add(2 * time.Hour),
			HttpOnly: true,
		})

		// Set CSRF token in a cookie
		http.SetCookie(w , &http.Cookie{
			Name:	"csrf_token",
			Value:	csrfToken,
			Expires: time.Now().Add(2 * time.Hour),
			HttpOnly: false,
		})

		//Store tokens in the database
		user.SessionToken = sessionToken
		user.CSRFToken = csrfToken
		users[email] = user 



		fmt.Fprintln(w, "Login succesful")
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := Authorize(r); err != nil {
		er := http.StatusUnauthorized
		http.Error(w , "Unauthorized", er)
		return
	}

	// Clear cookie
	http.SetCookie(w , &http.Cookie{
		Name:		"session_token",
		Value: 		"",
		Expires: 	time.Now().Add(-time.Hour),
		HttpOnly: 	true,
	})
	http.SetCookie(w , &http.Cookie{
		Name:		"csrf_token",
		Value: 		"",
		Expires: 	time.Now().Add(-time.Hour),
		HttpOnly: 	false,
	})

	//Clear the tokens from the database
	email := r.FormValue("email")
	user, _ := users[email]
	user.SessionToken = ""
	user.CSRFToken = ""
	users[email] = user

	fmt.Fprintln(w , "Logged out successfully!")
}

func Protected(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w , "Invalid request method", er)
		return
	}

	if err := Authorize(r); err != nil {
		er := http.StatusUnauthorized
		http.Error(w, "Unauthorized", er)
		return
	}
	email := r.FormValue("email")
	fmt.Fprintf(w , "CSRF validation successful! welcome, %s", email)
}
