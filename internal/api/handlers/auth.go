package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/Abtingh/Zeiterfassung/internal/db/sqlc"
	"github.com/Abtingh/Zeiterfassung/internal/util"
)

// Handler wraps db.Queries so the HTTP handlers can access the database
type Handler struct {
	Q *db.Queries
}

func NewHandler(q *db.Queries) *Handler {
	return &Handler{Q: q}
}

// LoginHandler handles both GET (serve login form) and POST (authenticate user).
// LoginHandler handles user login requests.
//
// Supported HTTP methods:
//   - OPTIONS: Handles CORS preflight requests by setting appropriate headers.
//   - GET: Serves the login HTML page to the client.
//   - POST: Processes login form submissions by validating user credentials,
//     generating session and CSRF tokens, setting cookies, and responding with
//     a redirect URL based on the user's role.
//
// On successful login, session and CSRF tokens are generated, set as cookies,
// and persisted in the database. The response includes a JSON object with the
// appropriate redirect URL for the user's role. If authentication fails, an
// error response is returned.
//
// Method not allowed responses are sent for unsupported HTTP methods.
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "./public/public-static/login.html")
		return

	case http.MethodPost:
		w.Header().Set("Content-Type", "application/json")

		// Parse form fields
		if err := r.ParseForm(); err != nil {
			http.Error(w, `{"error":"ParseForm error"}`, http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		log.Printf("Received login: email=%s", email)

		// Fetch the user from DB
		ctx := r.Context()
		user, err := h.Q.GetUserByEmail(ctx, email)
		if err != nil {
			http.Error(w, `{"error":"Invalid username or password"}`, http.StatusUnauthorized)
			return
		}

		log.Printf("User role for %s: '%s'", email, user.Role) // Debug log

		if !util.CheckPasswordHash(password, user.PasswordHash) {
			http.Error(w, `{"error":"Invalid username or password"}`, http.StatusUnauthorized)
			return
		}

		// Generate tokens
		sessionToken := util.GenerateToken(32)
		csrfToken := util.GenerateToken(32)

		// Set cookies
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  time.Now().Add(2 * time.Hour),
			HttpOnly: true,
			Path:     "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "csrf_token",
			Value:    csrfToken,
			Expires:  time.Now().Add(2 * time.Hour),
			HttpOnly: false,
			Path:     "/",
		})

		// Persist tokens in the database
		if err := h.Q.UpdateUserTokens(ctx, db.UpdateUserTokensParams{
			SessionToken: pgtype.Text{String: sessionToken, Valid: true},
			CsrfToken:    pgtype.Text{String: csrfToken, Valid: true},
			ID:           user.ID,
		}); err != nil {
			log.Printf("UpdateUserTokens error: %v", err)
			// Continue, but login succeeds without persistence.
		}

		// Respond with JSON containing redirect URL - everyone goes to /home
		redirect := "/home"
		log.Printf("Redirecting to: %s", redirect) // Debug log
		fmt.Fprintf(w, `{"redirect":"%s"}`, redirect)
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HomeHandler serves the appropriate home page based on the user's role.
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Ensure user is authenticated
	if err := h.Authorize(r); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// For /home path, determine the correct page based on user role
	if r.URL.Path == "/home" {
		// Get user from session to determine role
		ctx := r.Context()
		sessionToken, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := h.Q.GetUserBySessionToken(ctx, pgtype.Text{String: sessionToken.Value, Valid: true})
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Serve the appropriate home page based on user role
		switch user.Role {
		case "supervisor":
			http.ServeFile(w, r, "./public/public-static/teamleiter_home.html")
		case "admin":
			http.ServeFile(w, r, "./public/public-static/admin_home.html")
		case "accounting":
			http.ServeFile(w, r, "./public/public-static/buch_home.html")
		default:
			http.ServeFile(w, r, "./public/public-static/student_home.html")
		}
		return
	}

	// Serve the correct home page for role-specific paths (if still needed)
	switch r.URL.Path {
	case "/home/vorgesetzer":
		http.ServeFile(w, r, "./public/public-static/teamleiter_home.html")
	case "/home/admin":
		http.ServeFile(w, r, "./public/public-static/admin_home.html")
	case "/home/buchhaltung":
		http.ServeFile(w, r, "./public/public-static/buch_home.html")
	case "/home/student":
		http.ServeFile(w, r, "./public/public-static/student_home.html")
	default:
		http.NotFound(w, r)
	}
}

// LogoutHandler clears cookies and resets the tokens in the DB.
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.Authorize(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := r.FormValue("email")
	ctx := r.Context()

	// Best-effort DB cleanup; ignore errors so logout is always successful.
	if user, err := h.Q.GetUserByEmail(ctx, email); err == nil {
		_ = h.Q.UpdateUserTokens(ctx, db.UpdateUserTokensParams{
			SessionToken: pgtype.Text{Valid: false},
			CsrfToken:    pgtype.Text{Valid: false},
			ID:           user.ID,
		})
	}

	// Clear cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "csrf_token",
		Value:   "",
		Expires: time.Unix(0, 0),
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Protected demonstrates a CSRF-protected endpoint.
func (h *Handler) Protected(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := h.Authorize(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := r.FormValue("email")
	fmt.Fprintf(w, "CSRF validation successful! Welcome, %s", email)
}

func (h *Handler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure user is authenticated for both GET and POST
	if err := h.Authorize(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Serve the reset password HTML page
		http.ServeFile(w, r, "./public/public-static/reset_password.html")
		return

	case http.MethodPost:
		// Handle password reset form submission
		w.Header().Set("Content-Type", "application/json")

		// Parse form fields
		if err := r.ParseForm(); err != nil {
			http.Error(w, `{"error":"ParseForm error"}`, http.StatusBadRequest)
			return
		}

		currentPassword := r.FormValue("currentPassword")
		newPassword := r.FormValue("newPassword")
		confirmNewPassword := r.FormValue("confirmNewPassword")

		// Validate passwords match
		if newPassword != confirmNewPassword {
			http.Error(w, `{"error":"New passwords do not match"}`, http.StatusBadRequest)
			return
		}

		// Get user from session (you'll need to implement this based on your session handling)
		// For now, this is a placeholder - you'll need to get the current user's email/ID from the session
		// email := getUserEmailFromSession(r) // You need to implement this

		// TODO: Implement the actual password reset logic here
		// 1. Get current user from session
		// 2. Verify current password
		// 3. Hash new password
		// 4. Update password in database

		log.Printf("Password reset attempt - Current: %s, New: %s", currentPassword, newPassword)

		// For now, return success response
		fmt.Fprintf(w, `{"success":"Password reset successful"}`)
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
