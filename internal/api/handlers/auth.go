package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/Abtingh/Zeiterfassung/internal/db/sqlc"
	"github.com/Abtingh/Zeiterfassung/internal/util"
)

// Handler wraps db.Queries so the HTTP handlers can access the database
// without relying on a global variable or an in‑memory map.
type Handler struct {
	Q *db.Queries
}

func NewHandler(q *db.Queries) *Handler {
	return &Handler{Q: q}
}

// LoginHandler handles both GET (serve login form) and POST (authenticate user).
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "./public/public-static/login.html")
		return

	case http.MethodPost:
		// Parse form fields
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("ParseForm() error: %v", err), http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		// Fetch the user from DB instead of map[string]User
		ctx := r.Context()
		user, err := h.Q.GetUserByEmail(ctx, email)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Invalid username or password", http.StatusUnauthorized)
				return
			}
			log.Printf("GetUserByEmail error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !util.CheckPasswordHash(password, user.PasswordHash) {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
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
		})
		http.SetCookie(w, &http.Cookie{
			Name:    "csrf_token",
			Value:   csrfToken,
			Expires: time.Now().Add(2 * time.Hour),
		})

		// Persist tokens in the database
		if err := h.Q.UpdateUserTokens(ctx, db.UpdateUserTokensParams{
			SessionToken: pgtype.Text{String: sessionToken, Valid: true},
			CsrfToken:    pgtype.Text{String: csrfToken, Valid: true},
			ID:           user.ID,
		}); err != nil {
			log.Printf("UpdateUserTokens error: %v", err)
			// We still continue, but login succeeds without persistence.
		}

		fmt.Fprintln(w, "Login successful")

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	// Best‑effort DB cleanup; ignore errors so logout is always successful.
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

	fmt.Fprintln(w, "Logged out successfully!")
}

// Protected demonstrates a CSRF‑protected endpoint.
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
