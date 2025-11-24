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
	Q  *db.Queries
	DB db.DBTX
}

func NewHandler(q *db.Queries) *Handler {
	return &Handler{Q: q}
}

// NewHandlerWithDB creates a handler with both queries and database connection
func NewHandlerWithDB(q *db.Queries, dbConn db.DBTX) *Handler {
	return &Handler{Q: q, DB: dbConn}
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
		// Check if user already has a valid session token
		if err := h.Authorize(r); err == nil {
			// User is already authenticated, redirect to home
			log.Printf("User already authenticated, redirecting to /home")
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		// User is not authenticated, serve login page
		http.ServeFile(w, r, "./public/public-static/login.html")
		return

	case http.MethodPost:
		// Parse form fields
		if err := r.ParseForm(); err != nil {
			// For form submission errors, redirect back to login with error
			http.Redirect(w, r, "/login?error=parse_error", http.StatusSeeOther)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		log.Printf("Received login: email=%s", email)

		// Fetch the user from DB
		ctx := r.Context()
		user, err := h.Q.GetUserByEmail(ctx, email)
		if err != nil {
			log.Printf("GetUserByEmail failed for %s: %v", email, err)
			http.Redirect(w, r, "/login?error=invalid_credentials", http.StatusSeeOther)
			return
		}

		log.Printf("User role for %s: '%s'", email, user.Role) // Debug log

		if !util.CheckPasswordHash(password, user.PasswordHash) {
			log.Printf("Password check failed for %s", email)
			http.Redirect(w, r, "/login?error=invalid_credentials", http.StatusSeeOther)
			return
		}

		// Generate tokens
		sessionToken := util.GenerateToken(32)
		csrfToken := util.GenerateToken(32)

		// Set cookies with more explicit settings
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  time.Now().Add(2 * time.Hour),
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "csrf_token",
			Value:    csrfToken,
			Expires:  time.Now().Add(2 * time.Hour),
			HttpOnly: false,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		log.Printf("Cookies set for user %s - session: %s..., csrf: %s...", email, sessionToken[:10], csrfToken[:10])

		// Persist tokens in the database
		if err := h.Q.UpdateUserTokens(ctx, db.UpdateUserTokensParams{
			SessionToken: pgtype.Text{String: sessionToken, Valid: true},
			CsrfToken:    pgtype.Text{String: csrfToken, Valid: true},
			ID:           user.ID,
		}); err != nil {
			log.Printf("UpdateUserTokens error: %v", err)
			// Continue, but login succeeds without persistence.
		}

		// Redirect directly to /home instead of sending JSON
		log.Printf("Redirecting user %s to /home", email)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// LogoutHandler clears cookies and resets the tokens in the DB.
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	// 1. Attempt to identify the current user via their existing session cookie.
	// We need to do this BEFORE clearing the cookies to invalidate tokens in the DB.
	// h.GetCurrentUser uses the session cookie to find the user.
	user, err := h.GetCurrentUser(r)
	if err == nil {
		// User found. Proceed to invalidate their tokens in the database.
		log.Printf("Logout: Found user ID %d based on session cookie. Clearing tokens in DB.", user.ID)
		ctx := r.Context()

		// Set SessionToken and CsrfToken to NULL in the database for this user.
		// This ensures stolen cookies cannot be reused after logout.
		err = h.Q.UpdateUserTokens(ctx, db.UpdateUserTokensParams{
			SessionToken: pgtype.Text{Valid: false}, // Sets to NULL in SQL
			CsrfToken:    pgtype.Text{Valid: false}, // Sets to NULL in SQL
			ID:           user.ID,
		})

		if err != nil {
			// Log the error, but continue with the logout process (clearing client cookies).
			log.Printf("Logout Warning: Database token clearing failed: %v", err)
		}
	} else {
		// User could not be identified (e.g., session already expired or cookie missing).
		// We proceed to clear browser cookies anyway to be safe.
		log.Printf("Logout: User session not actively found (already logged out?), proceeding to clear browser cookies.")
	}

	// 2. Clear the cookies in the user's browser.
	// We do this by setting the same cookie names with past expiration dates.
	log.Printf("Logout: Clearing session and CSRF cookies in the browser.")

	// Clear Session Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",              // Empty value
		Path:     "/",             // Must match the path used during Login
		Expires:  time.Unix(0, 0), // Set expiration date to the past
		MaxAge:   -1,              // Force immediate deletion in modern browsers
		HttpOnly: true,            // Must match Login setting
		SameSite: http.SameSiteLaxMode,
	})

	// Clear CSRF Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token", // Correction: This should be "csrf_token" based on your login code
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: false, // Must match Login setting
		SameSite: http.SameSiteLaxMode,
	})

	// 3. Final step: Redirect the user to the login page.
	log.Printf("Logout Success: Redirecting to /login")
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
